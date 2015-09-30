package utils

//this is utils function wrapper for redigo redis -- same as revel redis cache
import (
	"github.com/garyburd/redigo/redis"
	"github.com/revel/revel/cache"
	"time"
	"errors"
	_ "log"
	"encoding/json"
)

// Length of time to cache an item.
const (
	DEFAULT = time.Duration(0)
	FOREVER = time.Duration(-1)
)

var (
	ErrCacheMiss = errors.New("revel/cache: key not found.")
	ErrNotStored = errors.New("revel/cache: not stored.")
)

// Getter is an interface for getting / decoding an element from a cache.
type Getter interface {
	// Get the content associated with the given key. decoding it into the given
	// pointer.
	//
	// Returns:
	//   - nil if the value was successfully retrieved and ptrValue set
	//   - ErrCacheMiss if the value was not in the cache
	//   - an implementation specific error otherwise
	Get(key string, ptrValue interface{}) error
}

// Wraps the Redis client to meet the Cache interface.
type RedisCache struct {
	Pool              *redis.Pool
	DefaultExpiration time.Duration
}

func (c RedisCache) Set(key string, value interface{}, expires time.Duration) error {
	conn := c.Pool.Get()
	defer conn.Close()
	if err := c.invoke(conn.Do, key, value, expires); err != nil {
		return err
	}

	jsonKey := "j:" + key;
	valueString, _ := json.Marshal(value)
	jsonString := string(valueString)

	//also save json-stringify of the object to redis -- so that nodejs can take up the value
	if err := c.invoke(conn.Do, jsonKey, jsonString, expires); err != nil {
		return err
	}

	return nil
}

func (c RedisCache) Add(key string, value interface{}, expires time.Duration) error {
	conn := c.Pool.Get()
	defer conn.Close()
	existed, err := exists(conn, key)
	if err != nil {
		return err
	} else if existed {
		return ErrNotStored
	}
	return c.invoke(conn.Do, key, value, expires)
}

func (c RedisCache) Replace(key string, value interface{}, expires time.Duration) error {
	conn := c.Pool.Get()
	defer conn.Close()
	existed, err := exists(conn, key)
	if err != nil {
		return err
	} else if !existed {
		return ErrNotStored
	}
	err = c.invoke(conn.Do, key, value, expires)
	if value == nil {
		return ErrNotStored
	} else {
		return err
	}
}

func (c RedisCache) Get(key string, ptrValue interface{}) error {
	conn := c.Pool.Get()
	defer conn.Close()
	raw, err := conn.Do("GET", key)
	if err != nil {
		return err
	} else if raw == nil {
		return ErrCacheMiss
	}
	item, err := redis.Bytes(raw, err)
	if err != nil {
		return err
	}
	return cache.Deserialize(item, ptrValue)
}

func generalizeStringSlice(strs []string) []interface{} {
	ret := make([]interface{}, len(strs))
	for i, str := range strs {
		ret[i] = str
	}
	return ret
}

func (c RedisCache) GetMulti(keys ...string) (Getter, error) {
	conn := c.Pool.Get()
	defer conn.Close()

	items, err := redis.Values(conn.Do("MGET", generalizeStringSlice(keys)...))
	if err != nil {
		return nil, err
	} else if items == nil {
		return nil, ErrCacheMiss
	}

	m := make(map[string][]byte)
	for i, key := range keys {
		m[key] = nil
		if i < len(items) && items[i] != nil {
			s, ok := items[i].([]byte)
			if ok {
				m[key] = s
			}
		}
	}
	return RedisItemMapGetter(m), nil
}

func exists(conn redis.Conn, key string) (bool, error) {
	return redis.Bool(conn.Do("EXISTS", key))
}

func (c RedisCache) Delete(key string) error {
	conn := c.Pool.Get()
	defer conn.Close()
	existed, err := redis.Bool(conn.Do("DEL", key))
	if err == nil && !existed {
		err = ErrCacheMiss
	}
	return err
}

func (c RedisCache) Increment(key string, delta uint64) (uint64, error) {
	conn := c.Pool.Get()
	defer conn.Close()
	// Check for existance *before* increment as per the cache contract.
	// redis will auto create the key, and we don't want that. Since we need to do increment
	// ourselves instead of natively via INCRBY (redis doesn't support wrapping), we get the value
	// and do the exists check this way to minimize calls to Redis
	val, err := conn.Do("GET", key)
	if err != nil {
		return 0, err
	} else if val == nil {
		return 0, ErrCacheMiss
	}
	currentVal, err := redis.Int64(val, nil)
	if err != nil {
		return 0, err
	}
	var sum int64 = currentVal + int64(delta)
	_, err = conn.Do("SET", key, sum)
	if err != nil {
		return 0, err
	}
	return uint64(sum), nil
}

func (c RedisCache) Decrement(key string, delta uint64) (newValue uint64, err error) {
	conn := c.Pool.Get()
	defer conn.Close()
	// Check for existance *before* increment as per the cache contract.
	// redis will auto create the key, and we don't want that, hence the exists call
	existed, err := exists(conn, key)
	if err != nil {
		return 0, err
	} else if !existed {
		return 0, ErrCacheMiss
	}
	// Decrement contract says you can only go to 0
	// so we go fetch the value and if the delta is greater than the amount,
	// 0 out the value
	currentVal, err := redis.Int64(conn.Do("GET", key))
	if err != nil {
		return 0, err
	}
	if delta > uint64(currentVal) {
		tempint, err := redis.Int64(conn.Do("DECRBY", key, currentVal))
		return uint64(tempint), err
	}
	tempint, err := redis.Int64(conn.Do("DECRBY", key, delta))
	return uint64(tempint), err
}

func (c RedisCache) Flush() error {
	conn := c.Pool.Get()
	defer conn.Close()
	_, err := conn.Do("FLUSHALL")
	return err
}

func (c RedisCache) invoke(f func(string, ...interface{}) (interface{}, error),
	key string, value interface{}, expires time.Duration) error {

	switch expires {
	case DEFAULT:
		expires = c.DefaultExpiration
	case FOREVER:
		expires = time.Duration(0)
	}

	b, err := cache.Serialize(value)
	if err != nil {
		return err
	}
	conn := c.Pool.Get()
	defer conn.Close()
	if expires > 0 {
		_, err := f("SETEX", key, int32(expires/time.Second), b)
		return err
	} else {
		_, err := f("SET", key, b)
		return err
	}
}

// Implement a Getter on top of the returned item map.
type RedisItemMapGetter map[string][]byte

func (g RedisItemMapGetter) Get(key string, ptrValue interface{}) error {
	item, ok := g[key]
	if !ok {
		return ErrCacheMiss
	}
	return cache.Deserialize(item, ptrValue)
}


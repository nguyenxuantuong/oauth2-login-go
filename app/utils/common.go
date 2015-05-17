package utils

import (
	"crypto/md5"
	"encoding/hex"
	"math/rand"
)

func GetMD5Hash(text string) string {
	hasher := md5.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func RandSeq(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}


//choose first non empty string from the list of arguments
func ChooseFirstNonEmpty(args ... string) string{
	for i := range(args){
		if(len(args[i]) != 0){
			return args[i];
		}
	}

	//otherwise, return empty string
	return "";
}
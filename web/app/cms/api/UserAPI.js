import 'core-js/es6/promise';
import 'whatwg-fetch';

//get list of all users
export function getUsers(
    limit, offset,
    url = `/api/user/list?limit=${limit}&offset=${offset}`
) {
    //return a promise
    return fetch(url).then(response =>
        response.json().then(json => {
            if(json.status === "success") {
                return json.data;
            }
            else {
                //it should be catch by the store as failed message
                throw new Error(json.error);
            }
        })
    )
}

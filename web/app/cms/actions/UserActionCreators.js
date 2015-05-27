import { dispatchAsync } from '../AppDispatcher';
import {ActionTypes} from '../constants/AppConstants.js';
import * as UserAPI from '../api/UserAPI';
import UserStore from '../stores/UserStore';

//export the request function users
export function getUsers(limit, offset) {
    //dispatch the action
    dispatchAsync(UserAPI.getUsers(limit, offset), {
        request: ActionTypes.GET_USERS,
        success: ActionTypes.GET_USERS_SUCCESS,
        failure: ActionTypes.GET_USERS_ERROR
    }, {});
}
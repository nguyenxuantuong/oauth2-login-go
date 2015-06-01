import { register } from '../AppDispatcher';
import { createStore, mergeIntoBag, isInBag } from '../utils/StoreUtils';
import selectn from 'selectn';
import AppConstants from '../constants/AppConstants';

var _searchUserArray = [];
var _searchUserLoading = false;

const UserStore = createStore({
    getSearchUserArray() {
        return _searchUserArray;
    },
    getSearchUserLoading() {
        return _searchUserLoading;
    }
});

UserStore.dispatchToken = register(action => {
    let actionType = selectn('type', action);

    //depend on action type; get the data from action
    switch (actionType) {
        case AppConstants.ActionTypes.GET_USERS:
            _searchUserLoading = true;
            break;

        case AppConstants.ActionTypes.GET_USERS_SUCCESS:
            _searchUserArray = selectn('response.results', action) || [];
            _searchUserLoading = false;
        default:
            break;
    }

    //then emit the event to notify the changes
    UserStore.emitChange();
});

export default UserStore;

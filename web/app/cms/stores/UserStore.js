import { register } from '../AppDispatcher';
import { createStore, mergeIntoBag, isInBag } from '../utils/StoreUtils';
import selectn from 'selectn';

const _users = {};

const UserStore = createStore({
  contains(login, fields) {
    return isInBag(_users, login, fields);
  },

  get(login) {
    return _users[login];
  },

  getAll() {
    return _users
  }
});

UserStore.dispatchToken = register(action => {
  console.log("dispatch token", action);
  UserStore.emitChange();
  //const responseUsers = selectn('response.entities.users', action);
  //if (responseUsers) {
  //  mergeIntoBag(_users, responseUsers);
  //  UserStore.emitChange();
  //}
});

export default UserStore;

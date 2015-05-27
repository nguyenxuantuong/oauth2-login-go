import React, { PropTypes } from 'react';
import * as UserActionCreators from '../actions/UserActionCreators';
import UserStore from '../stores/UserStore';
import connectToStores from '../utils/connectToStores';

class UserSearchPage{
    constructor() {
        this.renderUserTable = this.renderUserTable.bind(this);
        this.handleChangePage = this.handleChangePage.bind(this);
    }

    componentWillMount() {
        //sending request using the current props
        UserActionCreators.getUsers(10, 0);
    }

    componentWillReceiveProps(nextProps) {
        console.log("NEXT:", nextProps);
    }

    render() {
        const { users } = this.props;
        console.log("DATA:", users);

        return (
            <div>
                {this.renderUserTable()}
            </div>
        );
    }

    renderUserTable() {
        const {
            isLoading: isLoading,
        } = this.props;

        return (
            <div>
                Nothing het yet
            </div>
        );
    }

    handleChangePage() {
        //TODO: load another user page
    }
}

//declare static propTypes
UserSearchPage.propTypes = {
    users: React.PropTypes.array
};

UserSearchPage = connectToStores([UserStore], props => ({
    users: UserStore.getAll()
}))(UserSearchPage);

module.exports = UserSearchPage;
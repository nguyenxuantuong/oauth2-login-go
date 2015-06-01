import React, { PropTypes } from 'react';
import * as UserActionCreators from '../actions/UserActionCreators';
import UserStore from '../stores/UserStore';
import connectToStores from '../utils/connectToStores';
var Router = require('react-router');
var { Route, RouteHandler, Link } = Router;

class UserSearchPage {
    constructor() {
        this.renderUserTable = this.renderUserTable.bind(this);
        this.handleChangePage = this.handleChangePage.bind(this);
        this.pageSize = 10;
    }

    componentWillMount() {
        //sending request using the current props
        UserActionCreators.getUsers(this.pageSize, 0);
    }

    componentWillReceiveProps(nextProps) {
        //console.log("NEXT:", nextProps);
    }

    render() {
        const { users, isLoading } = this.props;

        return (
            <div className="container">
                <ul className="page-breadcrumb breadcrumb">
                    <li>
                        <Link to="dashboard">Dashboard</Link>
                        <i className="fa fa-circle"></i>
                    </li>
                    <li className="active">
                        Search Users
                    </li>
                </ul>

                <div className="row">
                    <div className="col-md-12">
                        <div className="portlet light">
                            <div className="portlet-title">
                                <div className="caption">
                                    <i className="icon-equalizer font-red-sunglo"></i>
                                    <span className="caption-subject bold uppercase">Users</span>
                                </div>

                                <div className="actions">
                                    <a className="btn btn-default green btn-circle">Add New User</a>
                                </div>
                            </div>
                            {!!isLoading &&
                            <div role="row" className="text-center">
                                <i className="fa fa-refresh fa-spin spinner-loading"></i>
                            </div>
                            }

                            {!isLoading && users.length == 0 &&
                            <div className="noEntry">
                                No User Found
                            </div>
                            }

                            {!isLoading && users.length >= 1 && this.renderUserTable()}
                        </div>
                    </div>
                </div>
            </div>
        );
    }

    renderUserTable() {
        const { users } = this.props;

        return (
            <div className="portlet-body">
                <div className="table-container">
                    <table className="table table-striped table-bordered table-hover">
                        <thead>
                        <tr role="row" className="heading">
                            <th className="table-header-title" width="10%">ID</th>
                            <th className="table-header-title">Name</th>
                            <th className="table-header-title" width="20%">UserName</th>
                            <th className="table-header-title" width="20%">Email</th>
                            <th className="table-header-title" width="14%">Status</th>
                            <th className="table-header-title" width="10%">Action</th>
                        </tr>
                        </thead>

                        <tbody>
                        {users.map((user, index) =>
                                <tr>
                                    <td>{user.id}</td>
                                    <td>{user.full_name}</td>
                                    <td>{user.user_name}</td>
                                    <td>{user.email}</td>
                                    <td>{user.status=='1'?"Activated":"Pending"}</td>
                                    <td className="text-center">
                                        <a className="btn btn-xs default btn-editable">
                                            <i className="fa fa-pencil"></i> Edit
                                        </a>
                                    </td>
                                </tr>
                        )}
                        </tbody>
                    </table>

                    <div>
                    </div>
                </div>
            </div>
        );
    }

    handleChangePage() {
        //TODO: load another user page
    }
}

//declare static propTypes
UserSearchPage.propTypes = {
    users: React.PropTypes.array,
    isLoading: React.PropTypes.bool
};

UserSearchPage = connectToStores([UserStore], props => ({
    users: UserStore.getSearchUserArray(),
    isLoading: UserStore.getSearchUserLoading()
}))(UserSearchPage);

module.exports = UserSearchPage;
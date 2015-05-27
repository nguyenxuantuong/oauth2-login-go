var React = require('react');

var AppConstants = require('../constants/AppConstants');
var Router = require('react-router');
var { Route, RouteHandler, Link } = Router;


class AppHeaderTop extends React.Component {
    constructor() {
        super();
    }

    render() {
        return (
            <div className="page-header-top">
                <div className="container">
                    <div className="page-logo">
                        <img src="/public/img/metronic-logo-light.png" alt="Oauth logo"/>
                    </div>
                    <a href="javascript:;" className="menu-toggler"></a>
                    <div className="top-menu">
                        <ul className="nav navbar-nav pull-right">
                            <li className="dropdown dropdown-user dropdown-dark">
                                <a href="javascript:;" className="dropdown-toggle" data-toggle="dropdown" data-hover="dropdown" data-close-others="true">
                                    <img alt="" className="img-circle" src="/public/img/blankavatar.png"/>
                                    <span className="username username-hide-mobile"></span>
                                </a>
                                <ul className="dropdown-menu dropdown-menu-default">
                                    <li>
                                        <a>
                                            <i className="icon-key"></i>
                                            Log Out
                                        </a>
                                    </li>
                                </ul>
                            </li>
                        </ul>
                    </div>
                </div>
            </div>
        )
    }
};

class AppHeaderMenu extends React.Component {
    render() {
       return (
           <div className="page-header-menu">
               <div className="container">
                   <div className="hor-menu ">
                       <ul className="nav navbar-nav">
                           <li className="active">
                               <Link to="dashboard">Dashboard</Link>
                           </li>

                           <li className="menu-dropdown classic-menu-dropdown ">
                               <a data-hover="megamenu-dropdown" data-close-others="true" data-toggle="dropdown" href="javascript:;">
                                   Create <i className="fa fa-angle-down"></i>
                               </a>
                               <ul className="dropdown-menu pull-left">
                                   <li className="dropdown">
                                       <Link to="createSSOClient">
                                           <i className="icon-briefcase"></i>
                                           SSO Client
                                       </Link>
                                   </li>
                               </ul>
                           </li>

                           <li className="menu-dropdown classic-menu-dropdown ">
                               <a data-hover="megamenu-dropdown" data-close-others="true" data-toggle="dropdown" href="javascript:;">
                                   Manage <i className="fa fa-angle-down"></i>
                               </a>
                               <ul className="dropdown-menu pull-left">
                                   <li className="dropdown">
                                       <Link to="searchSSOClients">
                                           <i className="icon-briefcase"></i>
                                           SSO Client
                                       </Link>
                                   </li>
                                   <li className="dropdown">
                                       <Link to="searchUsers">
                                           <i className="icon-briefcase"></i>
                                           Users
                                       </Link>
                                   </li>
                               </ul>
                           </li>
                       </ul>
                   </div>
               </div>
           </div>
       )
    }
}

//create main app
class App extends React.Component {
    render() {
        return (
            <div>
                <div className="page-header">
                    <AppHeaderTop />
                    <AppHeaderMenu />
                </div>

                <div className="page-content content-inner-page" xmlns="http://www.w3.org/1999/html">
                    <RouteHandler/>
                </div>
            </div>
        )
    }
}

module.exports = App;
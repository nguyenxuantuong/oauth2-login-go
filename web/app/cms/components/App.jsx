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
                        <img src="/public/img/logo-oauth-io-2.jpg" alt="Oauth logo"/>
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
                               <a>Dashboard</a>
                           </li>

                           <li className="menu-dropdown classic-menu-dropdown ">
                               <a data-hover="megamenu-dropdown" data-close-others="true" data-toggle="dropdown" href="javascript:;">
                                   Create <i className="fa fa-angle-down"></i>
                               </a>
                               <ul className="dropdown-menu pull-left">
                                   <li className="dropdown">
                                       <a>
                                           <i className="icon-briefcase"></i>
                                           SSO Client
                                       </a>
                                   </li>
                               </ul>
                           </li>

                           <li className="menu-dropdown classic-menu-dropdown ">
                               <a data-hover="megamenu-dropdown" data-close-others="true" data-toggle="dropdown" href="javascript:;">
                                   Manage <i className="fa fa-angle-down"></i>
                               </a>
                               <ul className="dropdown-menu pull-left">
                                   <li className="dropdown">
                                       <a>
                                           <i className="icon-briefcase"></i>
                                           SSO Client
                                       </a>
                                   </li>
                                   <li className="dropdown">
                                       <a>
                                           <i className="icon-briefcase"></i>
                                           Users
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

                <RouteHandler/>
            </div>
        )
    }
}

module.exports = App;
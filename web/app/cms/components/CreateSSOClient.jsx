var React = require('react');

var AppConstants = require('../constants/AppConstants');
var Router = require('react-router');
//var addons = require('react-addons');
//var ValidationMixin = require('react-validation-mixin');
var { Route, RouteHandler, Link } = Router;

class CreateContainer extends React.Component {
    render() {
        return (
            <div className="container">
                <ul className="page-breadcrumb breadcrumb">
                    <li>
                        <Link to="dashboard">Dashboard</Link>
                        <i className="fa fa-circle"></i>
                    </li>
                    <li className="active">
                        Create New SSO Client
                    </li>
                </ul>

                <div className="row">
                    <div className="col-md-12">

                        <div className="portlet light bordered">
                            <div className="portlet-title">
                                <div className="caption">
                                    <i className="icon-equalizer font-red-sunglo"></i>
                                    <span className="caption-subject font-red-sunglo bold uppercase">Create New SSO Client</span>
                                </div>
                                <div className="tools">
                                    <a href="" className="collapse" data-original-title="" title="">
                                    </a>
                                </div>
                            </div>
                            <div className="portlet-body form">
                                <form action="javascript:;" className="form-horizontal" name="addItemForm">
                                    <div className="form-actions">
                                        <div className="row">
                                            <div className="col-md-6">
                                                <div className="row">
                                                    <div className="col-md-offset-3 col-md-9">
                                                        <button type="submit" className="btn green">
                                                            <i className="icon-right-open-mini "></i>Save</button>
                                                    </div>
                                                </div>
                                            </div>
                                            <div className="col-md-6">

                                            </div>
                                        </div>
                                    </div>
                                </form>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        )
    }
}

class CreateSSOClient extends React.Component {
    render() {

    }
}

module.exports = CreateSSOClient;
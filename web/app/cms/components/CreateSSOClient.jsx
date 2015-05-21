var React = require('react');

var AppConstants = require('../constants/AppConstants');
var Router = require('react-router');
var { Route, RouteHandler, Link } = Router;

class CreateSSOClient extends React.Component {
    render() {
        return (
            <div>
                <h1>Create SSO Client</h1>
            </div>
        )
    }
}

module.exports = CreateSSOClient;
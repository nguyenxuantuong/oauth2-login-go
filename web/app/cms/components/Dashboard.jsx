var React = require('react');

var AppConstants = require('../constants/AppConstants');
var Router = require('react-router');
var { Route, RouteHandler, Link } = Router;

class Dashboard extends React.Component {
    render() {
        return (
            <div>
                <h1>Dashboard</h1>
            </div>
        )
    }
}

module.exports = Dashboard;
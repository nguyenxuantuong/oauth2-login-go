// For debugging in the browser
if (process.env.NODE_ENV !== 'production' &&
    require('react/lib/ExecutionEnvironment').canUseDOM) {
    window.React = require('react');
}

/**
 * Application Entry
 */
var ExecutionEnvironment = require('react/lib/ExecutionEnvironment');
var React = require('react');

var App = require('./components/App.jsx');
var Dashboard = require('./components/Dashboard.jsx');
var CreateSSOClient = require('./components/CreateSSOClient.jsx');

var AppActions = require('./actions/AppActions');
var AppConstants = require('./constants/AppConstants');

var LayoutConfig = AppConstants.LayoutConfig;

var Router = require('react-router');
var { Route, RouteHandler, Link } = Router;

//global route applications
var routes = (
    <Route handler={App}>
        <Route path="dashboard" handler={Dashboard}/>
        <Route path="ssoclient/create" handler={CreateSSOClient}/>
    </Route>
);

var Application = {
    start: function(bootstrap) {
        console.log("____APP START_____");

        // Ready the stores -- do nothing for now
        AppActions.initialize(bootstrap);

        // Client-side: mount the app component
        if (ExecutionEnvironment.canUseDOM) {
            var rootElement = document.getElementById(LayoutConfig.ROOT_ELEMENT_ID);
            //React.render(App(), rootElement);

            //render the app base on URL hash location
            Router.run(routes, Router.HashLocation, (Root) => {
                React.render(<Root/>, rootElement);
            });
        } else {
            // Server-side: return the app's html
            var rootComponentHTML = React.renderToString(App());
            return rootComponentHTML;
        }
    }
};

// Modules needed server-side
if (!ExecutionEnvironment.canUseDOM) {
    //Application.RouteUtils = require('./utils/RouteUtils');
}

module.exports = Application;

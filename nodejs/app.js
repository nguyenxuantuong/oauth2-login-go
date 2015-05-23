var koa = require('koa');
var hbs = require('koa-hbs');
var bodyParser = require('koa-bodyparser');
var http = require('http');
var session = require('koa-generic-session');
var redisStore = require('koa-redis');
var logger = require('koa-logger');

//Note: components using different react -- so need to require it from the web/ submodules
//otherwise, React will throw nasty error
var React   = require('./../web/node_modules/react');
var JSX     = require('node-jsx').install({extension: '.jsx'});

//some react component
var Login = require("./../web/app/components/login.jsx")
var Register = require("./../web/app/components/register.jsx")
var ForgotPassword = require("./../web/app/components/forgot-password.jsx")
var ResetPassword = require("./../web/app/components/reset-password.jsx")

//kao-router
var Router 		= require('koa-router');

var app = koa();

app.name = 'koa-oauth';

//some middlewares
app.use(bodyParser());
//using logger
app.use(logger())

//redis session store
app.use(session({
    store: redisStore()
}));

//static files
//TODO: move to nginx
app.use(require('koa-static')(__dirname + "/..", {maxAge: 3600000}));

// koa-hbs is middleware. `use` it before you want to render a view
app.use(hbs.middleware({
    viewPath: __dirname + '/views',
    defaultLayout: 'main',
    extname: '.hbs',
    partialsPath: __dirname + '/views/partials',
    layoutsPath: __dirname + '/views/layouts',
    disableCache: true //TODO: disable it in production enviroment
}));

//declare a router and adding routes
var router = new Router();
router.get("/home", function*(){
    yield this.render('auth', {title: 'koa-oauth'});
})

router.get("/render/login", function*(){
    var markup = React.renderToString(
        React.createElement(Login, {})
    );

    this.body = {
        success: true,
        data: markup
    };
})

router.get("/render/register", function*(){
    var markup = React.renderToString(
        React.createElement(Register, {})
    );

    this.body = {
        success: true,
        data: markup
    };
})

router.get("/render/resetPassword", function*(){
    var markup = React.renderToString(
        React.createElement(ResetPassword, {})
    );

    this.body = {
        success: true,
        data: markup
    };
})

router.get("/render/forgotPassword", function*(){
    var markup = React.renderToString(
        React.createElement(ForgotPassword, {})
    );

    this.body = {
        success: true,
        data: markup
    };
})

//main routes + handlers
app.use(router.middleware());

//listen for event
app.listen(3000);
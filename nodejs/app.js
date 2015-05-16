var koa = require('koa');
var hbs = require('koa-hbs');
var bodyParser = require('koa-bodyparser');
var http = require('http');
var session = require('koa-generic-session');
var redisStore = require('koa-redis');
var logger = require('koa-logger')

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

// koa-hbs is middleware. `use` it before you want to render a view
app.use(hbs.middleware({
    viewPath: __dirname + '/views',
    defaultLayout: 'main',
    extname: '.hbs',
    partialsPath: __dirname + '/views/partials',
    layoutsPath: __dirname + '/views/layouts',
    disableCache: true //TODO: disable it in production enviroment
}));

//main routes + handlers
app.use(function *(){
    yield this.render('auth', {title: 'koa-oauth'});
});

//listen for event
app.listen(3000);
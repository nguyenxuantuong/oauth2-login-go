//TODO:
Configuration = null;

var loginApp = angular.module('login', [
    "ui.router",
    'ui.bootstrap',
    'ui.keypress',
    'facebook',
    'googleplus'
]);

//some const auth-event; not being used much now
var AUTH_EVENTS = {
    forbidden: 'auth:FORBIDDEN',
    loginSuccess: 'auth:LOGIN_SUCCESS',
    loginFailed: 'auth:LOGIN_FAILED',
    logout: 'auth:LOGOUT',
    redirectEnded: 'auth:REDIRECT_ENDED'
};

loginApp.constant('AUTH_EVENTS', AUTH_EVENTS);

loginApp.config(['FacebookProvider', '$stateProvider', '$urlRouterProvider', '$locationProvider', 'GooglePlusProvider', '$httpProvider',
    function (FacebookProvider, $stateProvider, $urlRouterProvider, $locationProvider, GooglePlusProvider, $httpProvider) {
        $httpProvider.defaults.useXDomain = true;
        $httpProvider.defaults.withCredentials = true;
        delete $httpProvider.defaults.headers.common['X-Requested-With'];

        //configuration for the OATH2
        if(Configuration)
        {
            FacebookProvider.init(Configuration.oath.Facebook.clientId);
            GooglePlusProvider.init({
                clientId: Configuration.oath.Google.clientId,
                scopes: [
                    'https://www.googleapis.com/auth/plus.login',
                    'https://www.googleapis.com/auth/plus.profile.emails.read'
                ]
            });
        }
        
        //config router
        $stateProvider
            .state('login', {
                url: '/login',
                templateUrl: 'public/modules/login/login.tpl.html',
                controller: 'LoginController',
                resolve: {}
            })
            .state('register', {
                url: '/register',
                templateUrl: 'public/modules/login/register.tpl.html',
                controller: 'RegisterController',
                resolve: {}
            })
            .state("forgotPassword", {
                url: "/forgotPassword",
                templateUrl: "public/modules/login/passwordReset.tpl.html",
                controller: 'ForgotPasswordController',
                resolve: {}
            })
            .state("activateAccount", {
                url: "/activateAccount",
                templateUrl: "public/modules/login/activateAccount.tpl.html",
                controller: 'ActivateAccountController',
                resolve: {}
            })
            .state("resetPassword", {
                url: "/resetPassword",
                templateUrl: "public/modules/login/passwordReset.tpl.html",
                controller: 'ResetPasswordController',
                resolve: {}
            });

        $urlRouterProvider.otherwise("/login");
    }]);

loginApp.run(["$rootScope", "$state", "AUTH_EVENTS", "$q", "$window",
    function ($rootScope, $state, AUTH_EVENTS, $q, $window){
        //login successfully
        $rootScope.$on(AUTH_EVENTS.loginSuccess, function () {
            $window.location.href = "/";
        });
    }]);

loginApp.controller("LoginController", ['$rootScope', '$scope', '$q', '$http', 'Facebook', 'GooglePlus', '$location', function($rootScope, $scope, $q, $http, Facebook, GooglePlus, $location){
    $scope.userInput = {};

    var isLogin = false;

    $scope.login = function()
    {
        //avoid keypress twice
        if(!!isLogin)
        {
            return;
        }

        isLogin = true;

        setTimeout(function(){
            isLogin = false;
        }, 1300);

        //append whatever query string in the url to post request
        $http.post("/api/user/login", {
            email: $scope.userInput.email,
            password: CryptoJS.MD5($scope.userInput.password).toString()})
            .success(function(response) {
                var data = response;
                
                if(data.error)
                {
                    return showError("Error happen " + data.error)   
                }

                //if logged in successfully + it's authorize request
                if(location.search.indexOf("client_id") != -1 && location.search.indexOf("response_type") != -1){
                    //redirect it to the authorize API
                    window.location = window.location.origin + "/api/oath/authorize" + location.search;
                }
                
                $rootScope.user = data;
                //$rootScope.$broadcast(AUTH_EVENTS.loginSuccess, data);
            })
            .error(function(error){
                showError(error.stack || error.error || error);
            });
    };

    $scope.loginFacebook = function()
    {
        if(!!isLogin)
        {
            return;
        }

        isLogin = true;

        setTimeout(function(){
            isLogin = false;
        }, 1300);
        
        var that = this;

        if(!Facebook.isReady())
        {
            return console.error("facebook failed to initialize");
        }

        var accessToken;

        Facebook.login(function(){}, {scope: 'email'})
            .then(function(response){
                if(!response || !response.authResponse || !response.authResponse.accessToken || response.status != 'connected')
                {
                    return $q.reject({
                        error: "Facebook login failed"
                    });
                }

                accessToken = response.authResponse.accessToken;

                return $http.post("/api/user/facebookLogin", {fbId: response.authResponse.userID, accessToken: accessToken});
            })
            .then(function(response){
                var data = response;

                $rootScope.user = data;
                $rootScope.$broadcast(AUTH_EVENTS.loginSuccess, data);
            },function (error) {
                showError(error.stack || error.error || error);
            });
    };

    $scope.loginGoogle = function(){
        var that = this;
        var accessToken;

        if(!!isLogin)
        {
            return;
        }

        isLogin = true;

        setTimeout(function(){
            isLogin = false;
        }, 1300);
        
        GooglePlus.login()
            .then(function (authResult) {
                if(!authResult.access_token)
                {
                    return $q.reject({
                        error: "Google Login Failed"
                    });
                }

                accessToken = authResult.access_token;
                return GooglePlus.getUser();
            })
            .then(function(profile){
                var googleId = profile.id;
                return $http.post("/api/user/googleLogin", {googleId: googleId, accessToken: accessToken});
            })
            .then(function(response){
                var data = response;

                $rootScope.user = data;
                $rootScope.$broadcast(AUTH_EVENTS.loginSuccess, data);

            },function (error) {
                showError(error.stack || error.error || error);
            });
    };
}]);


loginApp.controller("RegisterController", ['$rootScope', '$scope', '$q', '$http', 'Facebook', 'GooglePlus', '$location', '$window', function($rootScope, $scope, $q, $http, Facebook, GooglePlus, $location, $window){
    $scope.userInput = {};

    $scope.register = function()
    {
        var that = this;
        $http.post("/api/user/register",{
            full_name: $scope.userInput.fullname,
            user_name: $scope.userInput.username,
            email: $scope.userInput.email,
            password: CryptoJS.MD5($scope.userInput.password).toString()
        })
            .success(function(response, status) {
                var data = response;
                return;
                if(data.emailActivation)
                {
                    $location.path("/activateAccount");
                }
                else
                {
                    $window.location.href = "index.html";
                }
            })
            .error(function(error){
                showError(error.stack || error.error || error);
            });
    };

    $scope.registerWithFacebook = function()
    {
        var that = this;

        if(!Facebook.isReady())
        {
            return console.error("facebook failed to initialize");
        }

        var accessToken;

        Facebook.login(function(){}, {scope: 'email'})
            .then(function(response){
                if(!response || !response.authResponse || !response.authResponse.accessToken || response.status != 'connected')
                {
                    return $q.reject({
                        error: "Facebook login failed"
                    });
                }

                accessToken = response.authResponse.accessToken;

                return Facebook.api('/me', function(){});
            })
            .then(function(response){
                var email = response.email,
                    fbId = response.id,
                    name = response.name;

                return $http.post("/api/user/facebookRegister", {
                    name: name,
                    email: email,
                    fbId: fbId,
                    accessToken: accessToken
                });
            })
            .then(function(response){
                var data = response;

                $rootScope.user = data;
                $rootScope.$broadcast(AUTH_EVENTS.loginSuccess, data);
            }, function(error){
                showError(error.stack || error.error || error);
            });
    };

    $scope.registerWithGoogle = function(){
        var that = this;
        var accessToken;

        GooglePlus.login()
            .then(function (authResult) {
                if(!authResult.access_token)
                {
                    return $q.reject({
                        error: "Google Login Failed"
                    });
                }

                accessToken = authResult.access_token;

                return GooglePlus.getUser();
            })
            .then(function(profile){
                var email = profile.email,
                    googleId = profile.id,
                    name = profile.name;

                return $http.post("/api/user/googleRegister", {
                    name: name,
                    email: email,
                    googleId: googleId,
                    accessToken: accessToken
                });
            })
            .then(function(response){
                var data = response;

                $rootScope.user = data;
                $rootScope.$broadcast(AUTH_EVENTS.loginSuccess, data);
            }, function(error){
                showError(error.stack || error.error || error);
            });
    };
}]);

loginApp.controller("ActivateAccountController", ['$rootScope', '$scope', '$q', '$http', 'Facebook', 'GooglePlus', '$location', function($rootScope, $scope, $q, $http, Facebook, GooglePlus, $location){
    var activationKey = $scope.activationKey = $location.search().activationKey;

    if(!activationKey) {
        return;
    }

    $http.post("/api/user/activateAccount", {activationKey: activationKey})
        .success(function(response, status) {
            console.log("activate successful");
            $scope.activationSuccess = true;
        })
        .error(function(error){
            if(error && error.error)
            {
                $scope.activationSuccess = false;
            }
            else
            {
                showError(error.stack || error.error || error);
            }
        });
}]);

loginApp.controller("ForgotPasswordController", ['$rootScope', '$scope', '$q', '$http', 'Facebook', 'GooglePlus', '$location', function($rootScope, $scope, $q, $http, Facebook, GooglePlus, $location){
    var that = this;
    $scope.userInput = {};
    $scope.passwordResetEmailSent = false;
    $scope.formSwitch = 'forgot-password';
    
    $scope.requestResetPassword = function()
    {
        var that = this;

        $http.post("/api/user/forgotPassword", {email: $scope.userInput.email})
            .success(function(response, status) {
                console.log("email sent successful");
                $scope.passwordResetEmailSent = true;
            })
            .error(function(error){
                if(error && error.error)
                {
                    $scope.activationSuccess = false;
                }
                else
                {
                    showError(error.stack || error.error || error);
                }
            });
    };
}]);

loginApp.controller("ResetPasswordController", ['$rootScope', '$scope', '$q', '$http', 'Facebook', 'GooglePlus', '$location', function($rootScope, $scope, $q, $http, Facebook, GooglePlus, $location){
    var that = this;
    $scope.userInput = {};
    var passwordResetKey = $scope.passwordResetKey = $location.search().passwordResetKey;
    $scope.updatedPassword = false;

    $scope.updatePassword = function()
    {
        var that = this;

        $http.post("/api/user/resetPassword", {
            passwordResetKey: passwordResetKey,
            password: CryptoJS.MD5($scope.userInput.password).toString()
        })
            .success(function(response, status) {
                var data = response;

                $scope.updatedPassword = true;
                $scope.updateStatus = data;
            })
            .error(function(error){
                showError(error.stack || error.error || error);
            });
    };
}]);
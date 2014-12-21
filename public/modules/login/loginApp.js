angular.module( 'mainApp', [
        'ui.router',
        'login'
    ])

    .config(['$stateProvider', '$urlRouterProvider', '$locationProvider', function myAppConfig ( $stateProvider, $urlRouterProvider, $locationProvider) {
        $urlRouterProvider.otherwise( '/login' );
    }])

    .run(['$http', '$rootScope', function run ($http, $rootScope) {
    }])

    .controller( 'mainCtrl', ['$scope', '$location', function AppCtrl ( $scope, $location ) {

    }]);
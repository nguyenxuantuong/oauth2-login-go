module.exports = function ( grunt ) {
    require('load-grunt-tasks')(grunt, {scope: 'devDependencies'});
    require('time-grunt')(grunt);

    //set pkg to package.json files
    grunt.config.set("pkg", grunt.file.readJSON("package.json"));

    //init config
    grunt.initConfig({
        sass: {                              // Task
            dist: {                            // Target
                options: {                       // Target options
                    style: 'expanded'
                },
                files: {                         // Dictionary of files
                    'css/login-main.css': 'css/login-main.scss'       // 'destination': 'source'
                }
            }
        }
    });

    //using grunt-sass which using fast C++ version: libsass
    grunt.loadNpmTasks('grunt-sass');
    //grunt.loadNpmTasks('grunt-contrib-sass');

    //register some task
    grunt.registerTask('default', ['sass']);
    grunt.registerTask( 'watch', [ 'build', 'delta' ] );
};

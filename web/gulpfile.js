var gulp = require('gulp');
var browserify = require('browserify');
var babelify = require('babelify');
var source = require('vinyl-source-stream');
var uglify = require('gulp-uglify');

//NOTE: We will enable some of experimental ES7 syntax support
gulp.task('build', function () {
    browserify({
        entries: './app/cms/app.js',
        extensions: ['.js'],
        debug: true,
        standalone: "Application"
    })
        .transform(babelify.configure({
            stage: 1
        }))
        .bundle()
        .pipe(source('cms.js'))
        .pipe(gulp.dest('../public/js/'));
});

//compress using gulp
gulp.task('compress', function(){
    return gulp.src('../public/js/cms.js')
        .pipe(uglify())
        .pipe(gulp.dest('../public/js/'));
});

//register some tasks
gulp.task('default', ['build']);
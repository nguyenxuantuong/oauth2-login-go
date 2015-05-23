var gulp = require('gulp');
var browserify = require('browserify');
var babelify = require('babelify');
var source = require('vinyl-source-stream');
var uglify = require('gulp-uglify');

gulp.task('build', function () {
    browserify({
        entries: './app/cms/app.js',
        extensions: ['.js'],
        debug: true,
        standalone: "Application"
    })
        .transform(babelify)
        .bundle()
        .pipe(source('cms.js'))
        .pipe(gulp.dest('../public/js/'));
});

gulp.task('compress', function(){
    return gulp.src('../public/js/cms.js')
        .pipe(uglify())
        .pipe(gulp.dest('../public/js/'));
})

gulp.task('default', ['build']);
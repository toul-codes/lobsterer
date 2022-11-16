'usr strict';

var gulp = require('gulp'),
    sass = require('gulp-sass'),
    sourcemaps = require('gulp-sourcemaps');
    livereload = require('gulp-livereload');

gulp.task('sass', function () {
  gulp.src('sass/**/*.scss')
    .pipe(sourcemaps.init())
    .pipe(sass({
      errLogToConsole: true
    }).on('error', sass.logError))
    .pipe(sourcemaps.write())
    .pipe(gulp.dest('public/css'))
    .pipe(livereload());
});

gulp.task('watch', function () {
  livereload.listen({ start: true });
  gulp.watch('sass/*.scss', ['sass']);
});

gulp.task('default', ['watch', 'sass']);
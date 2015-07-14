// Include gulp
var gulp = require('gulp'); 

// Include Our Plugins
var concat = require('gulp-concat');
var uglify = require('gulp-uglify');
var rename = require('gulp-rename');
var rename = require("gulp-rename");


gulp.task('scripts', function() {
  return gulp.src(['bower_components/jquery/dist/jquery.js',
		   'bower_components/raphael/raphael.js',
		   'bower_components/underscore/underscore.js',
		   'bower_components/js-sequence-diagrams/build/sequence-diagram-min.js'])
    .pipe(concat('sec.min.js'))
    .pipe(uglify())
    .pipe(rename("js/sec-doc.js"))
    .pipe(gulp.dest('static'));
});

gulp.task('default', ['scripts']);

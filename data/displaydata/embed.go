package displaydata

import (
	"embed"
)

// go:embed bootstrap-5.3.0-alpha3-dist/css/bootstrap-grid.min.css
// go:embed bootstrap-5.3.0-alpha3-dist/css/bootstrap-grid.min.css.map
// go:embed bootstrap-5.3.0-alpha3-dist/css/bootstrap-reboot.min.css
// go:embed bootstrap-5.3.0-alpha3-dist/css/bootstrap-reboot.min.css.map
// go:embed bootstrap-5.3.0-alpha3-dist/css/bootstrap-utilities.min.css
// go:embed bootstrap-5.3.0-alpha3-dist/css/bootstrap-utilities.min.css.map
// go:embed bootstrap-5.3.0-alpha3-dist/js/bootstrap.min.js
// go:embed bootstrap-5.3.0-alpha3-dist/js/bootstrap.min.js.map
//
//go:embed bootstrap-5.3.0-alpha3-dist/css/bootstrap.min.css
//go:embed bootstrap-5.3.0-alpha3-dist/css/bootstrap.min.css.map
//go:embed bootstrap-5.3.0-alpha3-dist/js/bootstrap.bundle.min.js
//go:embed bootstrap-5.3.0-alpha3-dist/js/bootstrap.bundle.min.js.map
//go:embed AdminLTE3.2/dist/css/adminlte.min.css
//go:embed AdminLTE3.2/dist/css/adminlte.min.css.map
//go:embed AdminLTE3.2/dist/js/adminlte.min.js
//go:embed AdminLTE3.2/dist/js/adminlte.min.js.map
//go:embed AdminLTE3.2/dist/img/*
//go:embed AdminLTE3.2/plugins/fontawesome-free/css/all.min.css
//go:embed AdminLTE3.2/plugins/fontawesome-free/webfonts/*
//go:embed AdminLTE3.2/plugins/jquery/jquery.min.js
//go:embed AdminLTE3.2/plugins/jquery/jquery.min.map
//go:embed css/sidebar.css
var WebRoot embed.FS

//go:embed templates/dashboard.gohtml
var TemplateRoot embed.FS

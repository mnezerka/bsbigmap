var map, base;

function addmap() {
    base = {
        {{range .}}
        '{{.Name}}': L.tileLayer('{{.Url}}', {
            name: '{{.Name}}', minZoom: {{.MinZoom}}, maxZoom: {{.MaxZoom}},
            attribution: '{{.Attribution}}'
            {{ if .SubDomains }}
            ,subdomains: '{{.SubDomains}}'
            {{ end }}
        }),
        {{end}}
    };
    map = L.map('map').setView([52, 11], 3);
    var control = L.control.layers(base);
    map.addControl(control);
    map.addLayer(base.mapycz);
}

// from http://wiki.openstreetmap.org/wiki/Slippy_map_tilenames
function lon2tile(lon, zoom) { return (Math.floor((lon+180)/360*Math.pow(2,zoom))); }
function lat2tile(lat, zoom) { return (Math.floor((1-Math.log(Math.tan(lat*Math.PI/180) + 1/Math.cos(lat*Math.PI/180))/Math.PI)/2 *Math.pow(2,zoom))); }

function getdata(f) {
    var z = map.getZoom(), b = map.getBounds();
    f.elements['zoom'].value = z;
    f.elements['xmin'].value = lon2tile(b.getWest(), z);
    f.elements['xmax'].value = lon2tile(b.getEast(), z);
    f.elements['ymin'].value = lat2tile(b.getNorth(), z);
    f.elements['ymax'].value = lat2tile(b.getSouth(), z);
    var layers = '';
    map.eachLayer(function(l) {
        if( layers.length ) layers += '|';
        layers += l.options.name;
    });
    f.elements['provider'].value = layers;
}

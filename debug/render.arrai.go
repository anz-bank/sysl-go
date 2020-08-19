package debug

const renderScript = `
### ------------------------------------------------------------------------ ###
###  debug_style.arrai                                                       ###
### ------------------------------------------------------------------------ ###

let debug_style = $`+"`"+`
<link href="https://fonts.googleapis.com/css2?family=B612+Mono:wght@400&family=Roboto+Mono&display=swap" rel="stylesheet">
<style>
body {
    max-width: 900px;
    margin: 40px auto;
    padding: 10px 40px;

    font-family: "Helvetica Neue", Helvetica, Arial, sans-serif;

    background-color: #fdfdfd;
    border: 1px solid #e1e4e8;
    border-radius: 2px;
}

h1, h2, h3 {
    border-bottom: 1px solid #eaecef;
}

table {
    border-collapse: collapse;
}

td, th {
    padding: 4px 8px;
    border: thin solid grey;
}

td, pre {
    font-family: 'B612 Mono', 'Courier New', monospace;
    font-size: 0.9em;
}

.method {
    line-height: 20px;
    background-color: rgb(36, 143, 178);
    color: white;
    text-transform: uppercase;
    font-family: Montserrat, sans-serif;
    padding: 3px 10px;
}

.route {
    font-family: 'B612 Mono', 'Courier New', monospace;
}

.url {
    float: right;
    font-size: small;
    text-decoration: underline;
    color: grey;
    line-height: 0;
}

.status {
    color: red;
}
.status.2 {
    color: green;
}
</style>
`+"`"+`;let debug_style_arrai = 
debug_style
;

### ------------------------------------------------------------------------ ###
###  debug_index.arrai                                                       ###
### ------------------------------------------------------------------------ ###

let render = \metadata
    let traceIds = metadata.entries => .@item.request.headers('Traceid')(0)?:{};
    $`+"`"+`
        <!DOCTYPE html>
        <html lang="en">
        <head>
            <meta charset="UTF-8">
            <title>Debug Traces</title>
            ${debug_style_arrai}
        </head>
        <body>
        <h1>Debug Traces</h1>

        <ul>
        ${traceIds orderby . >> $`+"`"+`
            <li><a href="/-/trace/${.}">${.}</a></li>
        `+"`"+`::\i}
        </ul>
        </body>
        </html>
    `+"`"+`;let debug_index_arrai = 

(
    :render
)
;

### ------------------------------------------------------------------------ ###
###  table.arrai                                                             ###
### ------------------------------------------------------------------------ ###

# Formats data as tables.

# Returns data formatted as an HTML table.
let htmlTable = \data $`+"`"+`
    <table>
        <tr><th>Key</th><th>Value</th></tr>
        ${(data => \(:@, :@value)
            $`+"`"+`<tr><td>${@}</td><td>${@value}</td></tr>`+"`"+`
        ) orderby .::\i}
    </table>
`+"`"+`;

# Returns data formatted as a Markdown table.
let markdownTable = \data $`+"`"+`
    |Key|Value|
    ${(data => \(:@, :@value)
        $`+"`"+`|${@}|${@value}|`+"`"+`
    ) orderby .::\i}
`+"`"+`;let table_arrai = 

(
    html: htmlTable,
    markdown: markdownTable,
)
;

### ------------------------------------------------------------------------ ###
###  debug_subtrace.arrai                                                    ###
### ------------------------------------------------------------------------ ###

let id = \. //seq.sub("/", "_", //str.lower($`+"`"+`
    ${.serviceName}_${.request.method}_${.request.route}`+"`"+`));

let render = \. $`+"`"+`
    <div class="subtrace ${id(.)}" style="display: none">
    <p class="url">${.request.url}</p>
    <p>${.serviceName}:
        <span class="method">${.request.method}</span>
        <span class="route">${.request.route}</span>
    </p>

    <h2>Request</h2>
    <h3>Headers</h3>
    ${cond .request.headers?:{} {{}: {}, h: table_arrai.html(h)}}

    <h3>Body</h3>
    <pre>${.request.body?:{}}</pre>

    <h2>Response</h2>
    <p>
        Status:
        <span class="status ${.response.statusCode // 100} ${.response.statusCode}">
            ${.response.statusCode} ${.response.status}
        </span>
        (${.response.latency/1000000:.1f}ms)
    </p>

    <h3>Headers</h3>
    ${cond .response.headers?:{} {{}: {}, h: table_arrai.html(h)}}

    <h3>Body</h3>
    <pre>${cond .response.body?:{} {x: x, _: "(none)"}}</pre>
    </div>
`+"`"+`;let debug_subtrace_arrai = 

(
    :render,
)
;

### ------------------------------------------------------------------------ ###
###  debug_trace.arrai                                                       ###
### ------------------------------------------------------------------------ ###

let render = \metadata \traceId \sd $`+"`"+`
    <!DOCTYPE html>
    <html lang="en">
    <head>
    <meta charset="UTF-8">
    <title>Trace Details</title>
    ${debug_style_arrai}
    </head>
    <body>
    <p><a href="/-/trace">Back</a></p>

    <h1>Trace Details</h1>

    <p>Trace ID: ${traceId}</p>

    ${sd}

    ${metadata.entries where .@item.request.headers('Traceid')?:{} = traceId >>
         debug_subtrace_arrai.render(.)
    ::\i}

    <script>
    function display(href) {
    let [_, lhs, arrow, rhs] = href.replace(/%20/g, ' ').match(/#([^-<>]+)([-<>]+)(.+)/);
    lhs = lhs && lhs.trim();
    rhs = rhs && rhs.trim();
    arrow = arrow && arrow.trim();
    console.log(lhs, arrow, rhs);

    const tc = rhs.replace(/[ /]/g, '_').toLowerCase();
    console.log(tc);
    document.querySelectorAll('div.subtrace').forEach(d => d.setAttribute('style', 'display:none'));
    document.querySelectorAll('div.subtrace.'+tc).forEach(d => d.setAttribute('style', 'display:block'));
    }

    document.querySelectorAll('a').forEach(
    a => a.addEventListener('mouseover',
        e => {
        display(e.currentTarget.getAttribute('href'))
        document.querySelectorAll('a').forEach(a => a.removeAttribute('style'))
        e.currentTarget.setAttribute('style', 'font-weight: bold');
        }
    )
    )
    </script>
    </body>
    </html>
`+"`"+`;let debug_trace_arrai = 

(
    :render,
)
;

### ------------------------------------------------------------------------ ###
###  json.arrai                                                              ###
### ------------------------------------------------------------------------ ###

# Helper functions for working with JSON data.

# Returns a "simplified" structure, replacing decoded tuples with their internal values.
# Intended for cases in which the types of null values is not interesting.
let rec simplify = \j
    cond j {
        [...]: j >> simplify(.),
        (:a): simplify(a),
        (:s): simplify(s),
        (): {},
        _: cond {
            j < 99999999999999999999: j,
            _: //log.print(j) >> simplify(.),
        },
    }
;let json_arrai = 

(
    :simplify,
)
;

### ------------------------------------------------------------------------ ###
###  debug.arrai                                                             ###
### ------------------------------------------------------------------------ ###

# Renders the sysl-go debug trace details screen.

let cleanHeaders = \hs //log.print(hs) >> cond . {[x]: x, _: .};
let cleanEntry = \ej 
    let clean = (
        serviceName: ej('serviceName'),
        request: //tuple(ej('request')),
        response: //tuple(ej('response')),
    );
    //log.print(clean) -> . +> (
        request: .request +> (headers: cleanHeaders(.request.headers)),
        response: .response +> (headers: cleanHeaders(.response.headers)),
    )
;
let cleanMetadata = \mj //tuple(json_arrai.simplify(mj)) -> \m
    m +> (entries: m.entries?:[] >> cleanEntry(.));
let parseMetadata = \mj cleanMetadata(//encoding.json.decode(mj));

(
    index: \m
        debug_index_arrai.render(parseMetadata(m)),
    trace: \m \traceId \svg
        debug_trace_arrai.render(parseMetadata(m), traceId, svg),
)
`

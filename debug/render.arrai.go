package debug

const renderScript = `
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
###  debug_subtrace.arrai                                                    ###
### ------------------------------------------------------------------------ ###

let id = \. //seq.sub("/", "_", //str.lower($`+"`"+`
    ${.serviceName}_${.request.method}_${.request.route}`+"`"+`));

let render = \. $`+"`"+`
    <div class="subtrace ${id(.)}" style="display: none">
    <p>${.serviceName}: ${.request.method} ${.request.route} (${.request.url})</p>

    <h2>Request</h2>
    <h3>Headers</h3>
    <pre>${.request.headers?:{}}</pre>

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
    <pre>${.response.headers?:{}}</pre>

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
    <style>
        .status {
        color: red;
        }
        .status.2 {
        color: green;
        }
    </style>
    </head>
    <body>
    <p><a href="/-/trace">Back</a></p>

    <h1>Trace Details</h1>

    <p>Trace ID: ${traceId}</p>

    ${sd}

    ${metadata.entries where .@item.request.headers('Traceid')(0)?:{} = traceId >>
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
        _: cond {
            j < 99999999999999999999: j,
            _: j >> simplify(.),
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

let cleanEntry = \ej (
    serviceName: ej('serviceName'),
    request: //tuple(ej('request')),
    response: //tuple(ej('response')),
);
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

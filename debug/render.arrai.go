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
    <pre>${.request.headers}</pre>

    <h3>Body</h3>
    <pre>${.request.body}</pre>

    <h2>Response</h2>
    <p>
        Status:
        <span class="status ${.response.statusCode // 100} ${.response.statusCode}">
            ${.response.statusCode} ${.response.status}
        </span>
        (${.response.latency/1000000:.1f}ms)
    </p>

    <h3>Headers</h3>
    <pre>${.response.headers}</pre>

    <h3>Body</h3>
    <pre>${cond .response.body {x: x, _: "(none)"}}</pre>
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
###  svg_grammar.arrai                                                       ###
### ------------------------------------------------------------------------ ###

# A grammar to parse SVG documents.

let g = {://grammar.lang.wbnf:
doc         -> header? node;
header      -> "<?xml" [^?]* "?>";
node        -> "<" tag=name attr* ("/>" | (">" (node | comment | text)* "</" name ">"));
name        -> [-:\w]+;
attr        -> name "=" '"' value=[^""]* '"';
comment     -> "<!--" comment_rest;
comment_rest -> "-->" | ([^-]+ | [-]) comment_rest;
text        -> [^<]+;

thisisntusedanywhere -> "<";
.wrapRE     -> /{\s*()\s*};
:};let svg_grammar_arrai = 

g
;

### ------------------------------------------------------------------------ ###
###  util.arrai                                                              ###
### ------------------------------------------------------------------------ ###

# A collection of helper functions for arr.ai.
#
# If generally useful, these should gradually migrate to a more standard library.

# Invokes a macro on a string as if it were source code at parsing time.
let invokeMacro = \macro \s
    macro -> (//dict(.@transform) >>> \rule \fn 
        fn(//grammar.parse(.@grammar, rule, s))).@value
;

# Transforms an AST into a simple tuple of its values.
# Useful for the @transform of a flat grammar.
let simpleTransform = \ast
    let d = //dict(ast) >> \term cond term {('':value, ...): value, _: {}};
    //tuple(d where .@value)
;

# Filters the nodes of a hierarchical data structure based on a (key, value) predicate.
# Key-value pairs for which the predicate returns false will be removed from the result.
let rec filterTree = \pred \ast
    cond ast {
        {(@:..., @value:...), ...}: ast where pred(.@, .@value) >> filterTree(pred, .),
        [...]: ast >> filterTree(pred, .),
        {...}: ast => filterTree(pred, .),
        (...): safetuple(//dict(ast) where pred(.@, .@value) >> filterTree(pred, .)),
        _: ast,
    }
;

# Sequentially applies `+"`"+`fn(accumulator, i)`+"`"+` for each `+"`"+`i`+"`"+` in `+"`"+`arr`+"`"+`. The `+"`"+`accumulator`+"`"+` is initialised
# to `+"`"+`val`+"`"+`, and updated to the result of `+"`"+`fn`+"`"+` after each invocation.
# Returns the final accumulated value.
let rec reduce = \arr \fn \val
    cond arr {
        [head, ...]:
            let tail = -1\(arr without (@:0, @item:head));
            reduce(tail, fn, fn(val, head)),
        _: val,
    }
;

# Performs `+"`"+`reduce`+"`"+` once on `+"`"+`arr`+"`"+`, and once for each array output of `+"`"+`fn`+"`"+`. Accumulates to the same
# value across all invocations.
let reduceFlat = \arr \fn \val
    reduce(arr, \z \i reduce(i, fn, z), val)
;

# Returns a sequence with any offset and holes removed.
let ranked = \s s rank (:.@);

# Workaround for https://github.com/arr-ai/arrai/issues/571
let safetuple = \d
    let rest = //tuple(d where .@ != '');
    cond d where .@ = '' {
        {(@:'', @value: value)}: rest +> (@: value),
        _: rest,
    };
# Explore constructs a dependency graph by starting at source and calling step
# to find adjacent nodes. Deps is the graph constructed so far.
# Self-edges are ignored.
let rec _explore = \source \step \deps
    cond {
        {source} & (deps => .@): deps,
        _:
            let next = step(source) where . != source;
            let deps = deps | {(@:source, @value: next)};
            reduce(next orderby ., \v \i _explore(i, step, v), deps)
    };
let explore = \source \step _explore(source, step, {});

# Unimported returns the set of nodes with no in-edges.
let unimported = \g (g => .@) where !({.} & //rel.union(g => .@value));

# Topsort returns an array of nodes in graph in dependency order.
let rec _topsort = \graph \sorted \sources
    cond sources orderby . {
            []: sorted,
            [..., tail]: 
                let adjs = graph(tail);
                let graph = graph where .@ != tail;
                let sources = (sources &~ {tail}) | (adjs & unimported(graph));
                _topsort(graph, sorted ++ [tail], sources)
        };
let topsort = \graph _topsort(graph, [], unimported(graph));let util_arrai = 

(
    :explore,
    :filterTree,
    :invokeMacro,
    :ranked,
    :reduce,
    :reduceFlat,
    :safetuple,
    :simpleTransform,
    :topsort,
    :unimported,
)
;

### ------------------------------------------------------------------------ ###
###  svg.arrai                                                               ###
### ------------------------------------------------------------------------ ###

# Functions for working with SVG documents.

let (:ranked, ...) = util_arrai;

let comment = \k \v k = "comment";
let at = \k \v //seq.has_prefix("@", k);
# Filters out nodes of an AST that are keyed by "comment" or "@*".
let pred = \k \v !comment(k, v) && !at(k, v);

# SVG attributes that have numeric values.
let nums = {'x', 'y', 'x1', 'x2', 'y1', 'y2', 'cx', 'cy', 'rx', 'ry',
    'textLength', 'font-size'};

# Transforms an SVG AST into a more natural arr.ai structure.
let transformDoc = \ast
    let rec transformNode = \node (
        @tag: node.tag.name.'',
        attrs: node.attr?:{} => \(@item:a, ...)
            let @ = ranked(a.name.'');
            let v = a.value.'';
            (:@, @value: cond {{@} & nums: //eval.value(v), _: v}),
        text: //seq.join("", node.text?:{} >> .''),
        children: node.node?:{} >> transformNode(.),
    );
    (header: //seq.join(' ', ast.header?.'':{}), root: transformNode(ast.node))
;

# Serializes an SVG model to SVG XML.
let toSvg = \m
    let attrToString = \as $`+"`"+`${as => $`+"`"+`${.@}="${.@value}"`+"`"+` orderby .:: }`+"`"+`;
    let rec toString = \n
        cond {
            n.children?:{}: $`+"`"+`
                <${n.@tag} ${attrToString(n.attrs)}>
                    ${n.children >> toString(.)::\i}${n.text}
                </${n.@tag}>
            `+"`"+`,
            _: $`+"`"+`<${n.@tag} ${attrToString(n.attrs)}>${n.text}</${n.@tag}>`+"`"+`,
        }
    ;

    $`+"`"+`
        ${m.header?:''}
        ${toString(m.root)}
    `+"`"+`
;

# Manipulation functions

# Returns the ranked tag of the node.
let tag = \node ranked(node.@tag);
# Returns the first g node.
let g = \svg ranked(svg.root.children where tag(.@item) = 'g')(0);
# Returns all nodes with the given tag.
let byTag = \svg \t (g(svg).children where tag(.@item) = t) => .@item;

# Returns the uppermost y coordinate of the node.
let getY = \n 
    let py = \poly //eval.value(//seq.split(',', poly.attrs('points'))(1));
    n.attrs('y')?:{} || n.attrs('y1')?:{} || py(n)
;

# Performs rough decoding of URL-encoded strings.
let urldecode = \in //seq.sub('&gt;', '>', //seq.sub('&lt;', '<', //seq.sub('%20', ' ', in)));

# Returns the LHS, RHS and arrow of an endpoint expression (e.g. x -> y).
let parts = \in
    let ing = //re.compile('#?\\s*([^-<>]*)\\s*([-<>]+)\\s*(.+)');
    let [[_, lhs, arrow, rhs]] = ing.match(//str.lower(urldecode(in)));
    (
        lhs: ranked(//seq.trim_suffix(' ', lhs)),
        arrow: ranked(arrow), 
        rhs: (//seq.trim_suffix(' ', rhs))
    )
;

let kids = \ns ns => .children => .@item;

# Returns text nodes that contain the given text. May be nested in a nodes.
let texts = \svg \text
    let raw = byTag(svg, 'text') where ranked(.text) = ranked(text);
    let as = kids(byTag(svg, 'a') where parts(.attrs('href')) = parts(text));
    let ats = kids(byTag(svg, 'a')) where ranked(.text) = ranked(text);
    raw | as | ats
;

# Returns the nodes most closely associated with the given text.
let byText = \svg \text
    let texts = texts(svg, text);
    //rel.union({'line', 'polygon'} => \t
        let elts = byTag(svg, t) orderby getY(.);
        texts => getY(.) => \y ranked(elts where getY(.@item) > y)(0)?:{}
    ) | texts where .
;

# Returns the SVG with the text colored.
let colored = \svg \color \nodes
    let tags = nodes => .@tag;
    let strokeRe = //re.compile('stroke: [^;]+');
    let strokeWidthRe = //re.compile('stroke-width: [^;]+');
    svg +> (root: svg.root +> (children: svg.root.children >>
        cond tag(.) {'g': . +> (children: .children >>
            cond {{.@tag} & tags: . +> (attrs: .attrs >>> \k \v
                cond k {
                    'style': strokeWidthRe.sub('stroke-width: 2.0', strokeRe.sub($`+"`"+`stroke: ${color}; cursor: pointer`+"`"+`, v)),
                    'fill': color,
                    _: v,
                }),
                _: .,
            }),
        _: .,
    }))
;let svg_arrai = 

(
    macro: (
        @grammar: svg_grammar_arrai,
        @transform: (doc: transformDoc),
    ),
    :toSvg,
    :byText,
    :colored,
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

let parseSvg = \svg util_arrai.invokeMacro(svg_arrai.macro, svg);

(
    index: \m
        debug_index_arrai.render(parseMetadata(m)),
    trace: \m \traceId \svg
        debug_trace_arrai.render(parseMetadata(m), traceId, svg),
)
`

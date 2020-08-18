package debug

import (
	"fmt"
	"github.com/arr-ai/arrai/syntax"
	"strings"
)

// UpdateSvg returns a copy of SVG with color applies to text and its associated elements.
func UpdateSvg(svg, texts, color string) (string, error) {
	f := strings.ReplaceAll(`(%s)("%s", "%s", "%s")`, `"`, "`")
	value, err := syntax.EvaluateExpr("", fmt.Sprintf(f, script, svg, texts, color))
	if err != nil {
		return "", err
	}

	return value.String(), nil
}

// The arr.ai script to perform the SVG update.
const script = `
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

# Sequentially applies ` + "`" + `fn(accumulator, i)` + "`" + ` for each ` + "`" + `i` + "`" + ` in ` + "`" + `arr` + "`" + `. The ` + "`" + `accumulator` + "`" + ` is initialised
# to ` + "`" + `val` + "`" + `, and updated to the result of ` + "`" + `fn` + "`" + ` after each invocation.
# Returns the final accumulated value.
let rec reduce = \arr \fn \val
    cond arr {
        [head, ...]:
            let tail = -1\(arr without (@:0, @item:head));
            reduce(tail, fn, fn(val, head)),
        _: val,
    }
;

# Performs ` + "`" + `reduce` + "`" + ` once on ` + "`" + `arr` + "`" + `, and once for each array output of ` + "`" + `fn` + "`" + `. Accumulates to the same
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

let svgGrammar = {://grammar.lang.wbnf:
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
:};

# Functions for working with SVG documents.

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
    let attrToString = \as $` + "`" + `${as => $` + "`" + `${.@}="${.@value}"` + "`" + ` orderby .:: }` + "`" + `;
    let rec toString = \n
        cond {
            n.children?:{}: $` + "`" + `
<${n.@tag} ${attrToString(n.attrs)}>
${n.children >> toString(.)::\i}${n.text}
</${n.@tag}>
` + "`" + `,
            _: $` + "`" + `<${n.@tag} ${attrToString(n.attrs)}>${n.text}</${n.@tag}>` + "`" + `,
        }
    ;

    $` + "`" + `
${m.header?:''}
${toString(m.root)}
` + "`" + `
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
                    'style': strokeWidthRe.sub('stroke-width: 2.0', strokeRe.sub($` + "`" + `stroke: ${color}; cursor: pointer` + "`" + `, v)),
                    'fill': color,
                    _: v,
                }),
                _: .,
            }),
        _: .,
    }))
;

let svgMacro = (@grammar: svgGrammar, @transform: (doc: transformDoc));


\sourceSvg \texts \color
    let model = invokeMacro(svgMacro, sourceSvg);
    let texts = //eval.value(texts);
    toSvg(
        colored(
            model,
            color,
            //rel.union(texts => byText(model, .))
        )
    )
`

+++
title = "Testing jQuery plugins with Jasmine"
aliases = ["/2011/05/testing-jquery-plugins-with-jasmine.html"]

[taxonomies]
tags = ["bdd", "jasmine", "javascript", "jquery"]
+++

Since I started working at [Globo.com](http://globo.com), I developed some
jQuery plugins (for internal use) with my team, and we are starting to test
these plugins using [Jasmine](https://pivotal.github.com/jasmine/), _“a
behavior-driven development framework for testing your JavaScript code”_. In
this post, I will show how to develop a very simple jQuery plugin (based on an
example that I learned with [Ricard D. Worth](http://rdworth.org/)): zebrafy.
This plugin “zebrafies” a table, applying different classes to odd and even
lines. Let’s start setting up a Jasmine environment... First step is download
the [standalone version of
Jasmine](https://pivotal.github.com/jasmine/download.html), then extract it and
edit the runner. The runner is a simple HTML file, that loads Jasmine and all
JavaScript files you want to test. But, wait... why not test using node.js or
something like this? Do I really need the browser on this test? You don’t
**need**, but I think it is important to test a plugin that works with the DOM
using a real browser. Let’s delete some files and lines from _SpecRunner.html_
file, so we adapt it for our plugin. This is how the structure is going to look
like:

```
.
├── SpecRunner.html
├── lib
│   ├── jasmine-1.0.2
│   │   ├── MIT.LICENSE
│   │   ├── jasmine-html.js
│   │   ├── jasmine.css
│   │   └── jasmine.js
│   └── jquery-1.6.1.min.js
├── spec
│   └── ZebrafySpec.js
└── src
    └── jquery.zebrafy.js
```

You can create the files `jquery.zebrafy.js` and `ZebrafySpec.js`, but
remember: it is BDD, we need to describe
the behavior first, then write the code. So let’s start writing the specs in
`ZebrafySpec.js` file using Jasmine. If you are familiar with
[RSpec](http://rspec.info) syntax, it’s easy to understand how to write spec
withs Jasmine, if you aren’t, here is the clue: Jasmine is a lib with some
functions used for writing tests in an easier way. I’m going to explain each
function “on demmand”, when we need something, we learn how to use it! ;)

First of all, we need to start a new test suite. Jasmine provides the
`describe` function for that, this function receives a string and another
function (a callback). The string describes the test suite and the function is
a callback that delimites the scope of the test suite. Here is the `Zebrafy`
suite:

```javascript
describe('Zebrafy', function () {

});
```

Let’s start describing the behavior we want to get from the plugin. The most
basic is: we want different CSS classes for odd an even lines in a table.
Jasmine provides the `it` function for writing the tests. It also receives a
string and a callback: the string is a description for the test and the
callback is the function executed as test. Here is the very first test:

```javascript
it('should apply classes zebrafy-odd and zebrafy-even to each other table lines', function () {
    var table = $("#zebra-table");
    table.zebrafy();
    expect(table).toBeZebrafyied();
});
```

Okay, here we go: in the first line of the callback, we are using jQuery to
select a table using the `#zebra-table` selector, which will look up for a
table with the ID attribute equals to _“zebra-table”_, but we don’t have this
table in the DOM. What about add a new table to the DOM in a hook executed
before the test run and remove the table in another hook that runs after the
test? Jasmine provide two functions: `beforeEach` and `afterEach`. Both
functions receive a callback function to be executed and, as the names suggest,
the `beforeEach` callback is called before each test run, and the `afterEach`
callback is called after the test run. Here are the hooks:

```javascript
beforeEach(function () {
    $('<table id="zebra-table"></table>').appendTo('body');
    for (var i=0; i < 10; i++) {
        $('<tr></tr>').append('<td></td>').append('<td></td>').append('<td></td>').appendTo('#zebra-table');
    };
});

afterEach(function () {
    $("#zebra-table").remove();
});
```

The `beforeEach` callback uses jQuery to create a table with 10 rows and 3
columns and add it to the DOM. In `afterEach` callback, we just remove that
table using jQuery again. Okay, now the table exists, let’s go back to the
test:

```javascript
it('should apply classes zebrafy-odd and zebrafy-even to each other table lines', function () {
    var table = $("#zebra-table");
    table.zebrafy();
    expect(table).toBeZebrafyied();
});
```

In the second line, we call our plugin, that is not ready yet, so let’s forward
to the next line, where we used the `expect` function. Jasmine provides this
function, that receives an object and executes a _matcher_ against it, there is
a lot of built-in matchers on Jasmine, but `toBeZebrafyied` is not a built-in
matcher. Here is where we know another Jasmine feature: the capability to write
custom matchers, but how to do this? You can call the `beforeEach` again, and
use the `addMatcher` method of Jasmine object:

```javascript
beforeEach(function () {
    this.addMatchers({
        toBeZebrafyied: function() {
            var isZebrafyied = true;

            this.actual.find("tr:even").each(function (index, tr) {
                isZebrafyied = $(tr).hasClass('zebrafy-odd') === false && $(tr).hasClass('zebrafy-even');
                if (!isZebrafyied) {
                    return;
                };
            });

            this.actual.find("tr:odd").each(function (index, tr) {
                isZebrafyied = $(tr).hasClass('zebrafy-odd') && $(tr).hasClass('zebrafy-even') === false;
                if (!isZebrafyied) {
                    return;
                };
            });

            return isZebrafyied;
        }
    });
});
```

The method `addMatchers` receives an object where each property is a matcher.
Your matcher can receive arguments if you want. The object being matched can be
accessed using `this.actual`, so here is what the method above does: it takes
all odd `<tr>` elements of the table (`this.actual`) and check if them have the
CSS class `zebrafy-odd` and don’t have the CSS class `zebrafy-even`, then do
the same checking with even `<tr>` lines.

Now that we have wrote the test, it’s time to write the plugin. Here some
jQuery code:

```javascript
(function ($) {
    $.fn.zebrafy = function () {
        this.find("tr:even").addClass("zebrafy-even");
        this.find("tr:odd").addClass("zebrafy-odd");
    };
})(jQuery);
```

I’m not going to explain [how to implement a jQuery
plugin](http://docs.jquery.com/Plugins/Authoring) neither [what are those
brackets on
function](http://benalman.com/news/2010/11/immediately-invoked-function-expression/),
this post aims to show how to use Jasmine to test jQuery plugins.

By convention, jQuery plugins are “chainable”, so let’s make sure the zebrafy
plugin is chainable using a spec:

```javascript
it('zebrafy should be chainable', function() {
    var table = $("#zebra-table");
    table.zebrafy().addClass('black-bg');
    expect(table.hasClass('black-bg')).toBeTruthy();
});
```

As you can see, we used the built-in matcher `toBeTruthy`, which asserts that
an object or expression is `true`. All we need to do is return the jQuery
object in the plugin and the test will pass:

```javascript
(function ($) {
    $.fn.zebrafy = function () {
        return this.each(function (index, table) {
            $(table).find("tr:even").addClass("zebrafy-even");
            $(table).find("tr:odd").addClass("zebrafy-odd");
        });
    };
})(jQuery);
```

So, the plugin is tested and ready to release! :) You can check the entire code
and test with more spec in a [GitHub
repository](https://github.com/fsouza/jquery-testing-jasmine).

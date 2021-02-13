+++
title = "Splinter: Python tool for acceptance tests on web applications"
aliases = ["/2011/05/splinter-python-tool-for-acceptance.html"]

[taxonomies]
tags = ["django", "python", "selenium", "web-development"]
+++

[Capybara](https://github.com/jnicklas/capybara) and
[Webrat](https://github.com/brynary/webrat) are great Ruby tools for acceptance
tests. A few months ago, we started a great tool for acceptance tests in
[Python](https://python.org/) web applications, called
[Splinter](http://splinter.rtfd.org). There are many acceptance test
tools on Python world: [Selenium](http://seleniumhq.org/),
[Alfajor](https://github.com/idealistdev/alfajor),
[Windmill](http://www.getwindmill.com/),
[Mechanize](http://wwwsearch.sourceforge.net/mechanize/),
[zope.testbrowser](https://pypi.python.org/pypi/zope.testbrowser), etc.
Splinter was not created to be another acceptance tool, but an abstract layer
over other tools, its goal is provide a unique API that make acceptance testing
easier and funnier.

In this post, I will show some basic usage of Splinter for simple web
application tests. [Splinter](https://github.com/cobrateam/splinter) is a tool
useful on tests of any web application. You can even test a Java web
application using Splinter. This post example is a "test" of a Facebook
feature, just because I want to focus on how to use Splinter, not on how to
write a web application. The feature to be tested is the creation of an event
(the Splinter sprint), following all the flow: first the user will login on
Facebook, then click on "Events" menu item, then click on "Create an Event"
button, enter all event information and click on "Create event" button. So,
let’s do it…

First step is create a
[Browser](https://splinter.readthedocs.io/en/latest/api/driver-and-element-api.html#module-splinter.browser)
instance, which will provide method for interactions with browser (where the
browser is: Firefox, Chrome, etc.). The code we need for it is very simple:

```python
browser = Browser("firefox")
```

`Browser` is a class and its constructor receives the driver to be used with
that instance. Nowadays, there are three drivers for Splinter: `firefox`,
`chrome` and `zope.testbrowser`. We are using Firefox, and you can easily use
Chrome by simply changing the driver from `firefox` to `chrome`. It’s also very
simple to add another driver to Splinter, and I plan to cover how to do that in
another blog post here.

A new browser session is started when we got the `browser` object, and this is
the object used for Firefox interactions. Let's start a new event on Facebook,
the Splinter Sprint. First of all, we need to _visit_ the Facebook homepage.
There is a `visit` method on Browser class, so we can use it:

```python
browser.visit("https://www.facebook.com")
```

`visit` is a blocking operation: it waits for page to load, then we can
navigate, click on links, fill forms, etc. Now we have Facebook homepage opened
on browser, and you probably know that we need to login on Facebook page, but
what if we are already logged in? So, let's create a method that login on
Facebook with provided authentication data only the user is not logged in
(imagine we are on a `Test`Case class):

```python
def do_login_if_need(self, username, password):
    if self.browser.is_element_present_by_css('div.menu_login_container'):
        self.browser.fill('email', username)
        self.browser.fill('pass', password)
        self.browser.find_by_css('div.menu_login_container input[type="submit"]').first.click()
        assert self.browser.is_element_present_by_css('li#navAccount')
```

What was made here? First of all, the method checks if there is an element
present on the page, using a CSS selector. It checks for a `div` that contains
the _username_ and _password_ fields. If that div is present, we tell the
browser object to fill those fields, then find the `submit` button and click on
it. The last line is an assert to guarantee that the login was successful and
the current page is the Facebook homepage (by checking the presence of
“Account” `li`).

We could also [find elements](http://splinter.rtfd.org/en/latest/finding.html)
by its texts, labels or whatever appears on screen, but remember: Facebook is
an internationalized web application, and we can’t test it using only a
specific language.

Okay, now we know how to visit a webpage, check if an element is present, fill
a form and click on a button. We're also logged in on Facebook and can finally
go ahead create the _Splinter sprint_ event. So, here is the event creation
flow, for a user:

1. On Facebook homepage, click on “Events” link, of left menu
1. The “Events” page will load, so click on “Create an Event” button
1. The user see a page with a form to create an event
1. Fill the date and chose the time
1. Define what is the name of the event, where it will happen and write a short
   description for it
1. Invite some guests
1. Upload a picture for the event
1. Click on “Create Event” button

We are going to do all these steps, except the 6th, because the Splinter Sprint
will just be a public event and we don’t need to invite anybody. There are some
boring AJAX requests on Facebook that we need to deal, so there is not only
Splinter code for those steps above. First step is click on “Events” link. All
we need to do is `find` the link and `click` on it:

```python
browser.find_by_css('li#navItem_events a').first.click()
```

The `find_by_css` method takes a CSS selector and returns an
[ElementList](http://splinter.rtfd.org/en/latest/api/element-list.html#splinter.element_list.ElementList).
So, we get the first element of the list (even when the selector returns only
an element, the return type is still a _list_) and click on it. Like `visit`
method, `click` is a blocking operation: the driver will only listen for new
actions when the request is finished (the page is loaded).

We’re finally on "new event" page, and there is a form on screen waiting for data of the _Splinter Sprint_. Let’s fill the form. Here is the code for it:

```python
browser.fill('event_startIntlDisplay', '5/21/2011')
browser.select('start_time_min', '480')
browser.fill('name', 'Splinter sprint')
browser.fill('location', 'Rio de Janeiro, Brazil')
browser.fill('desc', 'For more info, check out the #cobratem channel on freenode!')
```

That is it: the event is going to happen on May 21th 2011, at 8:00 in the
morning (480 minutes). As we know, the event name is _Splinter sprint_, and we
are going to join some guys down here in Brazil. We filled out the form using
`fill` and `select` methods.

The `fill` method is used to fill a "fillable" field (a textarea, an input,
etc.). It receives two strings: the first is the _name_ of the field to fill
and the second is the _value_ that will fill the field. `select` is used to
select an option in a select element (a “combo box”). It also receives two
string parameters: the first is the _name_ of the select element, and the
second is the _value_ of the option being selected.

Imagine you have the following select element:

```html
<select name="gender">
    <option value="m">Male</option>
    <option value="f">Female</option>
</select>
```

To select “Male”, you would call the select method this way:

```python
browser.select("gender", "m")
```

The last action before click on “Create Event” button is upload a picture for
the event. On new event page, Facebook loads the file field for picture
uploading inside an `iframe`, so we need to switch to this frame and interact
with the form present inside the frame. To show the frame, we need to click on
“Add Event Photo” button and then switch to it, we already know how click on a
link:

```python
browser.find_by_css('div.eventEditUpload a.uiButton').first.click()
```

When we click this link, Facebook makes an asynchronous request, which means
the driver does not stay blocked waiting the end of the request, so if we try
to interact with the frame BEFORE it appears, we will get an
`ElementDoesNotExist` exception. Splinter provides the `is_element_present`
method that receives an argument called `wait_time`, which is the time Splinter
will wait for the element to appear on the screen. If the element does not
appear on screen, we can’t go on, so we can assume the test failed (remember we
are testing a Facebook feature):

```python
if not browser.is_element_present_by_css('iframe#upload_pic_frame', wait_time=10):
    fail("The upload pic iframe did'n't appear :(")
```

The `is_element_present_by_css` method takes a CSS selector and tries to find
an element using it. It also receives a `wait_time` parameter that indicates a
time out for the search of the element. So, if the `iframe` element with
_ID=”upload_pic_frame”_ is not present or doesn’t appear in the screen after 10
seconds, the method returns `False`, otherwise it returns `True`.

> **Important:** `fail` is a pseudocode sample and doesn’t exist (if you’re
> using `unittest` library, you can invoke `self.fail` in a TestCase, exactly
> what I did in [complete snippet for this
> example](https://github.com/cobrateam/splinter/blob/master/samples/test_facebook_events.py
> "Snippet for creating a new event on Facebook using Splinter"), available at
> Github).

Now we see the `iframe` element on screen and we can finally upload the
picture. Imagine we have a variable that contains the path of the picture (and
not a file object, `StringIO`, or something like this), and this variable name
is `picture_path`, this is the code we need:

```python
with browser.get_iframe('upload_pic_frame') as frame:
    frame.attach_file('pic', picture_path)
    time.sleep(10)
```

Splinter provides the `get_iframe` method that changes the context and returns
another objet to interact with the content of the frame. So we call the
`attach_file` method, who also receives two strings: the first is the _name_ of
the input element and the second is the absolute _path_ to the file being sent.
Facebook also uploads the picture asynchronously, but there’s no way to wait
some element to appear on screen, so I just put Python to sleep 10 seconds on
last line.

After finish all these steps, we can finally click on “Create Event” button and
asserts that Facebook created it:

```python
browser.find_by_css('label.uiButton input[type="submit"]').first.click()
title = browser.find_by_css('h1 span').first.text
assert title == 'Splinter sprint'
```

After create an event, Facebook redirects the browser to the event page, so we
can check if it really happened by asserting the header of the page. That’s
what the code above does: in the new event page, it click on submit button, and
after the redirect, get the text of a span element and asserts that this text
equals to _“Splinter sprint”_.

That is it! This post was an overview on Splinter API. Check out the [complete
snippet](https://github.com/cobrateam/splinter/blob/master/samples/test_facebook_events.py),
written as a test case and also check out [Splinter repository at
Github](https://github.com/cobrateam/splinter).

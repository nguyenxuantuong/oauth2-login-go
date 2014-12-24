# Welcome to Auth

## Getting Started

Auth is based on Revel -- A high-productivity web framework for the [Go language](http://www.golang.org/).

### Start the web server:

    revel run auth

### Run test
    revel test auth test (TestSuite -- optional)

### Description of Contents

The default directory structure of a generated Revel application:

    myapp               App root
      app               App sources
        controllers     App controllers
          init.go       Interceptor registration
        models          App domain models
        routes          Reverse routes (generated code)
        views           Templates
      tests             Test suites
      conf              Configuration files
        app.conf        Main configuration file
        routes          Routes definition
      messages          Message files
      public            Public assets
        css             CSS files
        js              Javascript files
        images          Image files

app

    The app directory contains the source code and templates for your application.

conf

    The conf directory contains the application’s configuration files. There are two main configuration files:

    * app.conf, the main configuration file for the application, which contains standard configuration parameters
    * routes, the routes definition file.


messages

    The messages directory contains all localized message files.

public

    Resources stored in the public directory are static assets that are served directly by the Web server. Typically it is split into three standard sub-directories for images, CSS stylesheets and JavaScript files.

    The names of these directories may be anything; the developer need only update the routes.

test

    Tests are kept in the tests directory. Revel provides a testing framework that makes it easy to write and run functional tests against your application.


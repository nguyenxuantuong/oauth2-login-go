/**
 * Application Entry
 */
var ExecutionEnvironment = require('react/lib/ExecutionEnvironment');
var React = require('react');
var addons = require('react-addons');
var ValidationMixin = require('react-validation-mixin');
var Joi = require('joi');
var cx = require('react/lib/cx');
var Q = require("q");
var superagent = require("superagent");

var Login = React.createClass({
    mixins: [ValidationMixin, addons.LinkedStateMixin],
    validatorTypes:  {
        email: Joi.string().email().label('Email'),
        password: Joi.string().regex(/[a-zA-Z0-9]{3,30}/).label('Password')
    },
    getInitialState: function() {
        return {
            email: null,
            password: null,
            rememberMe: null
        };
    },
    componentDidMount: function() {

    },
    componentWillUnmount: function() {

    },
    renderHelpText: function(message) {
        return (
            <span className="help-block">{message}</span>
        );
    },
    getClasses: function(field) {
        return addons.classSet({
            'form-group': true,
            'has-error': !this.isValid(field)
        });
    },
    handleReset: function(event) {
        event.preventDefault();
        this.clearValidations();
        this.setState(this.getInitialState());
    },
    handleSubmit: function(event) {
        var that = this;

        event.preventDefault();
        var onValidate = function(error, validationErrors) {
            if (error) {
                this.setState({
                    feedback: 'Form is invalid do not submit'
                });
            } else {
                //now post to server to register
                console.log("Current state", this.state);

                Q.ninvoke(superagent.post("/api/user/login")
                    .send({
                        email: this.state.email,
                        password: this.state.password
                    })
                    .set('Accept', 'application/json'), "end")
                    .then(function(response){
                        var body = response.body;
                        if(body.status === "success"){
                            console.log("Register successfully", body.data);
                            if(body.data && body.data.redirect) {
                                return window.location = body.data.redirect;
                            }
                            else{
                                //redirect user into home page
                                window.location = "/home";
                            }
                        }
                        else
                        {
                            that.setState({feedback : body.errors || "Unable to login. Please try again later."});
                        }
                    })
            }
        }.bind(this);
        this.validate(onValidate);
    },
    render: function() {
        var that = this;

        return (
            <div>
                <form className="login-form" name="loginForm" method="post" onSubmit={this.handleSubmit}>
                    <h3 className="form-title primary-text bold">Login</h3>
                    <div className={cx({
                            'hidden': !that.state.feedback,
                            'alert':1, 'alert-danger': 1
                        })}>
                        <i className="fa fa-info-circle info"></i>
                        <span>{that.state.feedback}</span>
                    </div>

                    <div className="alert alert-danger display-hide">
                        <button className="close" data-close="alert"></button>
                        <span>
                            Enter any username and password.
                        </span>
                    </div>

                    <div className={this.getClasses('email')}>
                        <label className="control-label visible-ie8 visible-ie9">Email</label>
                        <input className="form-control placeholder-no-fix"
                               autofocus
                               id='email'
                               type="text" autocomplete="off"
                               placeholder="Enter your email" name="email"
                               valueLink={this.linkState('email')} onBlur={this.handleValidation('email')} />
                        {this.getValidationMessages('email').map(this.renderHelpText)}
                    </div>

                    <div className={this.getClasses('password')}>
                        <label className="control-label visible-ie8 visible-ie9">Password</label>
                        <input className="form-control placeholder-no-fix"
                               id="password"
                               type="password" autocomplete="off"
                               placeholder="Enter your password" name="password"
                               valueLink={this.linkState('password')} onBlur={this.handleValidation('password')}/>
                        {/* {this.getValidationMessages('password').map(this.renderHelpText)} */}
                        <span className={cx({
                            'hidden': this.getValidationMessages('password').length==0
                        })}>
                            {["\"Password\" is in incorrect format"].map(this.renderHelpText)}
                        </span>
                    </div>

                    <div className="form-actions">
                        <div style={{'margin-bottom': '21px', 'margin-top': '24px'}}>
                            <label className="rememberme check">
                                <input type="checkbox"
                                       style={{"margin-left": "-7px", "margin-right": "7px;"}}
                                       name="remember" valueLink={this.linkState('rememberMe')}/>
                                Keep me logged in</label>
                            <a href="/forgotPassword" id="forget-password" className="forget-password">Reset Password?</a>
                        </div>

                        <div>
                            <button type="submit" className="btn btn-main uppercase">Login</button>
                        </div>
                    </div>

                    <div className="login-options">
                        <h4>Or login with</h4>
                        <ul className="social-icons">
                            <li>
                                <a className="social-icon-color facebook" data-original-title="Facebook" href="/facebook/login"></a>
                            </li>
                            <li>
                                <a className="social-icon-color twitter" data-original-title="Twitter" href="/twitter/login"></a>
                            </li>
                            <li>
                                <a className="social-icon-color googleplus" data-original-title="Google Plus" href="/google/login"></a>
                            </li>
                        </ul>
                    </div>

                    <div className="create-account bottom-bar">
                        <a href="/register" id="register-btn" style={{"color": "white"}}
                           className="uppercase">Create an account</a>
                    </div>
                </form>
            </div>
        );
    }
});

//export the app
module.exports = Login;
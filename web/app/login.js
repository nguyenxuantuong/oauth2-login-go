// For debugging in the browser
if (process.env.NODE_ENV !== 'production' &&
    require('react/lib/ExecutionEnvironment').canUseDOM) {
    window.React = require('react');
}

/**
 * Application Entry
 */
var ExecutionEnvironment = require('react/lib/ExecutionEnvironment');
var React = require('react');
var addons = require('react-addons');
var ValidationMixin = require('react-validation-mixin');
var Joi = require('joi');
var cx = require('react/lib/cx');

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
        event.preventDefault();
        var onValidate = function(error, validationErrors) {
            if (error) {
                this.setState({
                    feedback: 'Form is invalid do not submit'
                });
            } else {
                //now post to server to register
                console.log("Current state", this.state);
            }
        }.bind(this);
        this.validate(onValidate);
    },
    render: function() {
        return (
            <div>
                <form className="login-form" name="loginForm" method="post" onSubmit={this.handleSubmit}>
                    <h3 className="form-title primary-text bold">Login</h3>
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
                            <a href="/passwordReset" id="forget-password" className="forget-password">Reset Password?</a>
                        </div>

                        <div>
                            <button type="submit" className="btn btn-main uppercase">Login</button>
                        </div>
                    </div>

                    <div className="login-options">
                        <h4>Or login with</h4>
                        <ul className="social-icons">
                            <li>
                                <a className="social-icon-color facebook" data-original-title="facebook" href="#"></a>
                            </li>
                            <li>
                                <a className="social-icon-color twitter" data-original-title="Twitter" href="#"></a>
                            </li>
                            <li>
                                <a className="social-icon-color googleplus" data-original-title="Goole Plus" href="#"></a>
                            </li>
                            <li>
                                <a className="social-icon-color linkedin" data-original-title="Linkedin" href="#"></a>
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

//for-now, always run in browser so it might be not necessary
if (ExecutionEnvironment.canUseDOM) {
    var rootElement = document.getElementById("react-root");
    React.render(Login(), rootElement);
}
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
var superagent = require("superagent");
var Q = require("q");

var Register = React.createClass({
    mixins: [ValidationMixin, addons.LinkedStateMixin],
    validatorTypes:  {
        fullName: Joi.string().required().label('Full Name'),
        userName:  Joi.string().alphanum().min(3).max(30).required().label('Username'),
        email: Joi.string().email().label('Email Address'),
        password: Joi.string().regex(/[a-zA-Z0-9]{3,30}/).label('Password'),
        verifyPassword: Joi.any().valid(Joi.ref('password')).required().label('Password Confirmation')
    },
    getInitialState: function() {
        return {
            fullName: null,
            userName: null,
            email: null,
            password: null,
            verifyPassword: null,
            feedback: null,
            agreeTerm: null
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

                Q.ninvoke(superagent.post("/api/user/register")
                    .send({
                        full_name: this.state.fullName,
                        user_name: this.state.userName,
                        email: this.state.email,
                        password: this.state.password
                    })
                    .set('Accept', 'application/json'), "end")
                .then(function(response){
                        var body = response.body;
                        if(body.status === "success"){
                            console.log("Register successfully", body.data);
                        }
                        else
                        {
                            that.setState({feedback : body.errors || "Unable to register new account. Please try again later."});
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
                    <h3 className="form-title primary-text bold">Register</h3>

                    <div className={cx({
                            'hidden': !that.state.feedback,
                            'alert':1, 'alert-danger': 1
                        })}>
                        <i className="fa fa-info-circle info"></i>
                        <span>{that.state.feedback}</span>
                    </div>

                    <div className="alert alert-danger display-hide">
                        <button className="close" data-close="alert"></button>
                        <span>Enter any username and password. </span>
                    </div>

                    <div className={this.getClasses('userName')}>
                        <label className="control-label visible-ie8 visible-ie9">Username</label>
                        <input className="form-control placeholder-no-fix"
                               autofocus
                               id="userName"
                               type="text" autocomplete="off"
                               placeholder="Enter your username" name="userName"
                               valueLink={this.linkState('userName')} onBlur={this.handleValidation('userName')}/>
                        {this.getValidationMessages('userName').map(this.renderHelpText)}
                    </div>

                    <div className={this.getClasses('email')}>
                        <label className="control-label visible-ie8 visible-ie9">Email</label>
                        <input className="form-control placeholder-no-fix"
                               type="text" autocomplete="off"
                               id="email"
                               placeholder="Enter your email" name="email"
                               valueLink={this.linkState('email')} onBlur={this.handleValidation('email')}/>
                        {this.getValidationMessages('email').map(this.renderHelpText)}
                    </div>

                    <div className={this.getClasses('fullName')}>
                        <label className="control-label visible-ie8 visible-ie9">Full Name</label>
                        <input className="form-control placeholder-no-fix"
                               type="text" autocomplete="off"
                               placeholder="Enter your name" name="fullname"
                               valueLink={this.linkState('fullName')} onBlur={this.handleValidation('fullName')}/>
                        {this.getValidationMessages('fullName').map(this.renderHelpText)}
                    </div>


                    <div className={this.getClasses('password')}>
                        <label className="control-label visible-ie8 visible-ie9">Password</label>
                        <input className="form-control placeholder-no-fix"
                               id="password"
                               ui-keypress="{13:'login($event)'}"
                               type="password" autocomplete="off"
                               placeholder="Enter your password" name="password"
                               valueLink={this.linkState('password')} onBlur={this.handleValidation('password')}/>
                        <span className={cx({
                            'hidden': this.getValidationMessages('password').length==0
                        })}>
                            {["\"Password\" is in incorrect format"].map(this.renderHelpText)}
                        </span>
                    </div>

                    <div className={this.getClasses('verifyPassword')}>
                        <label className="control-label visible-ie8 visible-ie9">Retype Password</label>
                        <input className="form-control placeholder-no-fix"
                               id="retype-password"
                               ui-keypress="{13:'login($event)'}"
                               type="password" autocomplete="off"
                               valueLink={this.linkState('verifyPassword')} onBlur={this.handleValidation('verifyPassword')}
                               placeholder="Retype password to confirm" name="password"/>
                        <span className={cx({
                            'hidden': this.getValidationMessages('verifyPassword').length==0
                        })}>
                            {["\"Password\" does not match"].map(this.renderHelpText)}
                        </span>
                    </div>

                    <div className="form-actions service-privacy">
                        <div className="agree-div">
                            <label className="rememberme check">
                                <input type="checkbox" className="agree-check-box"
                                       name="agreeTerm" valueLink={this.linkState('agreeTerm')}/>
                                I agree to the <a href="#">Terms of Service </a>&amp; <a href="#">Privacy Policy </a>
                            </label>
                        </div>

                        <div>
                            <button type="submit" className="btn btn-main uppercase">Register</button>
                        </div>
                    </div>

                    <div className="login-options">
                        <h4>Or register with</h4>
                        <ul className="social-icons">
                            <li>
                                <a className="social-icon-color facebook"
                                   data-original-title="facebook" href="#"></a>
                            </li>
                            <li>
                                <a className="social-icon-color twitter"
                                   data-original-title="Twitter" href="#"></a>
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
                        <a href="/login" id="register-btn"
                           className="uppercase">Login</a>
                    </div>
                </form>
            </div>
        );
    }
});

//for-now, always run in browser so it might be not necessary
if (ExecutionEnvironment.canUseDOM) {
    var rootElement = document.getElementById("react-root");
    React.render(Register(), rootElement);
}
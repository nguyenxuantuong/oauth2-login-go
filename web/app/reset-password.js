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
var pathToRegexp = require('path-to-regexp')

var ResetPassword = React.createClass({
    mixins: [ValidationMixin, addons.LinkedStateMixin],
    validatorTypes:  {
        password: Joi.string().regex(/[a-zA-Z0-9]{3,30}/).label('Password'),
        verifyPassword: Joi.any().valid(Joi.ref('password')).required().label('Password Confirmation')
    },
    getInitialState: function() {
        return {
            email: null,
            password: null,
            verifyPassword: null,
            resetSuccess: false
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
        var that = this;

        var onValidate = function(error, validationErrors) {
            if (error) {
                this.setState({
                    feedback: 'Form is invalid do not submit'
                });
            } else {
                //now post to server to register
                console.log("Current state", this.state);
                var keys = []
                var re = pathToRegexp('/resetPassword/:key', keys)

                var match = re.exec(window.location.pathname)
                if(match.length < 2){
                    return that.setState({feedback: "Missing reset password key"})
                }

                var resetKey = match[1].split("?")[0];

                console.log("Current state", this.state);

                Q.ninvoke(superagent.post("/api/user/resetPassword/" + resetKey)
                    .query({
                        newPassword: that.state.password
                    })
                    .set('Accept', 'application/json'), "end")
                    .then(function(response){
                        var body = response.body;
                        if(body.status === "success"){
                            console.log("Reset successfully", body.data);

                            //TODO: now redirect to the home page or any redirectUrl page; or just show successful page
                            that.setState({resetSuccess: true})
                        }
                        else
                        {
                            that.setState({feedback : body.errors || "Unable to activate the account. Please try again later."});
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
                <form name="resetPasswordForm" className="form-vertical forget-form" onSubmit={this.handleSubmit}>

                    <h3 className="primary-text bold">Reset Password ?</h3>

                    <div className={cx({
                            'hidden': !that.state.feedback || !!that.state.resetSuccess,
                            'alert':1, 'alert-danger': 1
                        })}>
                        <i className="fa fa-info-circle info"></i>
                        <span>{that.state.feedback}</span>
                    </div>

                    <div className={cx({
                            'hidden': !that.state.resetSuccess,
                            'row alert alert-info password-reset-sent': 1
                        })}>
                        <i className="fa fa-info-circle info"></i>
                        Your password has been successfully reset. Click
                        <a href="/login"><strong> here</strong></a> to go to the login page.
                    </div>

                    <span className={cx({
                            'hidden': !!that.state.resetSuccess
                        })}>
                        <p>
                            Please enter your new password.
                        </p>

                        <div className={this.getClasses('password')}>
                            <input className="form-control"
                                   type="password" autocomplete="off"
                                   placeholder="Password" name="password" required
                                   valueLink={this.linkState('password')} onBlur={this.handleValidation('password')} />
                             <span className={cx({
                            'hidden': this.getValidationMessages('password').length==0
                            })}>
                                {["\"Password\" is in incorrect format"].map(this.renderHelpText)}
                            </span>
                        </div>

                        <div className={this.getClasses('verifyPassword')}>
                            <input className="form-control"
                                   type="password" autocomplete="off"
                                   placeholder="Re-type Your Password"
                                   valueLink={this.linkState('verifyPassword')} onBlur={this.handleValidation('verifyPassword')}
                                   name="confirmPassword" required/>
                             <span className={cx({
                                'hidden': this.getValidationMessages('verifyPassword').length==0
                                 })}>
                                {["\"Password\" does not match"].map(this.renderHelpText)}
                             </span>
                        </div>

                        <div className="form-actions">
                            <a type="button" href="/login" className="btn btn-default">BACK</a>
                            <button type="submit"
                                    className="btn btn-main uppercase pull-right">Submit</button>
                        </div>
                    </span>

                    <div className="bottom-bar"> </div>
                </form>

            </div>
        );
    }
});

//for-now, always run in browser so it might be not necessary
if (ExecutionEnvironment.canUseDOM) {
    var rootElement = document.getElementById("react-root");
    React.render(ResetPassword(), rootElement);
}


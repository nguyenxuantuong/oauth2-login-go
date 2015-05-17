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

var ForgotPassword = React.createClass({
    mixins: [ValidationMixin, addons.LinkedStateMixin],
    validatorTypes:  {
        email: Joi.string().email().label('Email')
    },
    getInitialState: function() {
        return {
            email: null,
            emailSent: false
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

                Q.ninvoke(superagent.post("/api/user/requestPasswordReset")
                    .query({
                        email: this.state.email
                    })
                    .set('Accept', 'application/json'), "end")
                    .then(function(response){
                        var body = response.body;
                        if(body.status === "success"){
                            console.log("Register successfully", body.data);
                            that.setState({emailSent: true})
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
                <form name="forgotPasswordForm"
                      className="form-vertical forget-form"
                      method="post" onSubmit={this.handleSubmit}>
                    <h3 className="primary-text bold">Reset Password ?</h3>

                    <div className={cx({
                            'hidden': !that.state.feedback || that.state.emailSent,
                            'alert':1, 'alert-danger': 1
                        })}>
                        <i className="fa fa-info-circle info"></i>
                        <span>{that.state.feedback}</span>
                    </div>

                    <div className={cx({
                            'hidden': !that.state.emailSent,
                            'row alert alert-info password-reset-sent': 1
                        })}>
                        <i className="fa fa-info-circle info"></i>
                        An email has been sent to you. Please follow the instructions provided in the email to reset your password.
                    </div>

                    <div className={cx({
                            'hidden': !!that.state.emailSent
                        })}>
                        <p>
                            A link to reset your password will be sent there
                        </p>

                        <div className={this.getClasses('email')}>
                            <input className="form-control placeholder-no-fix"
                                   type="email"
                                   id='email'
                                   valueLink={this.linkState('email')} onBlur={this.handleValidation('email')}
                                   autocomplete="off" placeholder="Email to send password to"
                                   autofocus required
                                   name="email"/>
                            {this.getValidationMessages('email').map(this.renderHelpText)}
                        </div>

                        <div className="form-actions">
                            <a type="button" href="/login" id="back-btn" className="btn btn-default">BACK</a>
                            <button type="submit"
                                    className="btn btn-main uppercase pull-right">Submit</button>
                        </div>
                    </div>

                    <div className="bottom-bar"> </div>
                </form>
            </div>
        );
    }
});

module.exports = ForgotPassword;
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

var Register = React.createClass({
    getInitialState: function() {
        return {

        };
    },
    componentDidMount: function() {

    },
    componentWillUnmount: function() {

    },
    render: function() {
        return (
            <div>
                <form className="forget-form"
                      ng-show="formSwitch=='forgot-password'"
                      name="forgotPasswordForm"
                      className="form-vertical forget-form"
                      method="post">
                    <h3 className="primary-text bold">Reset Password ?</h3>

                    <div className="row alert alert-danger" ng-show="forgotPasswordForm.errorMessage">
                        <i className="fa fa-info-circle info"></i>
                        <span></span>
                    </div>

                    <div ng-show="forgotPasswordForm.passwordResetSent" className="row alert alert-info password-reset-sent">
                        <i className="fa fa-info-circle info"></i>
                        An email has been sent to you. Please follow the instructions provided in the email to reset your password.
                    </div>

                    <p>
                        A link to reset your password will be sent there
                    </p>
                    <div className="form-group">
                        <label className="text-danger help-inline help-small no-left-padding"
                               ng-show="forgotPasswordForm.submitted && forgotPasswordForm.email.$invalid">
                            The email is invalid.
                        </label>
                        <input className="form-control placeholder-no-fix"
                               type="email"
                               ng-model="resetEmail"
                               autocomplete="off" placeholder="Email to send password to"
                               autofocus required
                               name="email"/>
                    </div>

                    <div className="form-actions">
                        <a type="button" ui-sref="login" id="back-btn" className="btn btn-default">BACK</a>
                        <button type="submit"
                                ng-disabled="forgotPasswordForm.passwordResetSent"
                                className="btn btn-main uppercase pull-right">Submit</button>
                    </div>

                    <div className="bottom-bar"> </div>
                </form>

                <form ng-show="formSwitch=='reset-password'"
                      name="resetPasswordForm" className="form-vertical forget-form"
                      ng-submit="(resetPasswordForm.submitted=true) && resetPasswordForm.$valid && resetPassword(newPassword)">

                    <h3 className="primary-text bold">Reset Password ?</h3>

                    <div className="row alert alert-danger" ng-show="resetPasswordForm.errorMessage">
                        <i className="fa fa-info-circle info"></i>
                        <span ng-bind="resetPasswordForm.errorMessage"></span>
                    </div>

                    <div ng-show="resetPasswordForm.resetSuccess" className="row alert alert-info password-reset-sent">
                        <i className="fa fa-info-circle info"></i>
                        Your password has been successfully reset. Click
                        <a ui-sref="login"><strong>here</strong></a> to go to the login page.
                    </div>

                    <span ng-hide="!!resetPasswordForm.resetSuccess">
                        <p>
                            Please enter your new password.
                        </p>

                        <div className="form-group">
                            <label className="text-danger help-inline help-small no-left-padding"
                                   ng-show="resetPasswordForm.submitted && resetPasswordForm.password.$invalid">
                                Password must be at least 6 characters long.
                            </label>
                            <input className="form-control"
                                   type="password" autocomplete="off"
                                   placeholder="Password" name="password" required/>
                        </div>

                        <div className="form-group">
                            <label className="text-danger help-inline help-small no-left-padding"
                                   ng-show="resetPasswordForm.submitted && resetPasswordForm.confirmPassword.$invalid">
                                The passwords are different.
                            </label>
                            <input className="form-control"
                                   type="password" autocomplete="off"
                                   placeholder="Re-type Your Password"
                                   ng-model="confirmPassword"
                                   name="confirmPassword" required/>
                        </div>

                        <div className="form-actions">
                            <a type="button" ui-sref="login" className="btn btn-default">BACK</a>
                            <button type="submit"
                                    ng-disabled="resetPasswordForm.resetSuccess"
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
    React.render(Register(), rootElement);
}


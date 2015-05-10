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

var AccountActivation = React.createClass({
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
                <form className="form-vertical forget-form form-horizontal login account-activation-form"
                    name="accountActivationForm" method="post" role="form">
                    <h3 className="primary-text bold"> Account Activation</h3>

                    <div className="row alert alert-danger">
                        <i className="fa fa-info-circle info"></i>
                        <span ng-bind="accountActivationForm.errorMessage"></span>
                    </div>

                    <div ng-show="activationSuccess" className="row alert alert-info password-reset-sent">
                        <i className="fa fa-info-circle info"></i>
                        Your account has been successfully activated. Click
                        <a ui-sref="login"><strong>here</strong></a> to go to the login page.
                    </div>

                    <span>
                        <p>
                            Enter your desired password to activate
                        </p>

                        <div className="form-group">
                            <label className="text-danger help-inline help-small no-left-padding">
                                Password must be at least 6 characters long.
                            </label>

                            <input className="form-control"
                                   type="password" autocomplete="off"
                                   placeholder="Password"
                                   name="password" required/>
                        </div>

                        <div className="form-group">
                            <label className="text-danger help-inline help-small no-left-padding"
                                   ng-show="accountActivationForm.submitted && confirmPassword != newPassword">
                                The passwords are different.
                            </label>

                            <input className="form-control"
                                   type="password" autocomplete="off"
                                   placeholder="Re-type Your Password"
                                   ng-model="confirmPassword"
                                   name="confirmPassword" required/>
                        </div>

                        <div className="form-actions">
                            <button type="submit"
                                    className="btn btn-main uppercase pull-right">Activate</button>
                        </div>
                    </span>
                </form>
            </div>
        );
    }
});

//for-now, always run in browser so it might be not necessary
if (ExecutionEnvironment.canUseDOM) {
    var rootElement = document.getElementById("react-root");
    React.render(AccountActivation(), rootElement);
}
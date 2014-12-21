//this is just some utils, sugar functions
console.debug = function()
{
    if (typeof DEBUG === "undefined") return;
    if (!DEBUG) return;

    var a = (arguments[0] || "").grey;
    if (a) arguments[0] = a;
    console.log.apply(this, arguments);
}

function safeApply(scope, fn) {
    (scope.$$phase || (scope.$root && scope.$root.$$phase)) ? fn() : scope.$apply(fn);
}

function showError(title, error, options) {
    return bootbox.alert("ERROR: " + title || error);
}

function showPopup(title, text, time)
{
    $.gritter.add({
        title: title,
        text: text || " ",
        time: time || 2000
    });
}
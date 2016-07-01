$(function() {
    $("#name").val("");
    $("#name").keyup(function(e) {
        if ($(this).val() == "") {
            $("#name-holder").removeClass("has-success").addClass("has-warning");
        } else {
            $("#name-holder").removeClass("has-warning has-error").addClass("has-default");
        }
        if (e.keyCode == 13) {
            send();
        }
    })
    $("#send").click(function(e) {
        send();
    });
});

function send() {
    $("#alert").hide();
    var name = $("#name").val();
    if (name == "") {
        $("#name-holder").removeClass("has-warning has-default").addClass("has-error");
        $("#alert").show().text("Please provide a name.");
    } else {
        $.get(
                "/api/v0/list/"+name
             ).done(function(data) {
            $("#name").val("");
            var Parent = document.getElementById("table");
            while(Parent.hasChildNodes()) {
                Parent.removeChild(Parent.firstChild);
            }
                if (data == null) {
                    while(Parent.hasChildNodes()) {
                        Parent.removeChild(Parent.firstChild);
                    }
                    tr = $('<tr/>');
                    tr.append("<td>" + "user not found" + "</td>");
                    $('table').append(tr);
                }
                else {
                    for (var i in data) {
                        tr = $('<tr/>');
                        tr.append("<td>" + data[i].name + "</td>");
                        tr.append("<td>" + data[i].stargazers_count + "</td>");
                        $('table').append(tr);
                    }
                }
        }).fail(function(e) {
            var Parent = document.getElementById("table");
            while(Parent.hasChildNodes()) {
                Parent.removeChild(Parent.firstChild);
            }
            tr = $('<tr/>');
            tr.append("<td>" + "user not found" + "</td>");
            $('table').append(tr);
        });
    }
};

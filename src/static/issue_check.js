
var label_list = []

function checkedIssue(checkID) {
    console.log("checked: " + checkID)
    var box = document.getElementById(checkID);

    console.log("labels: " + label_list)

    if (box.checked == true) {
        label_list.push(checkID)
    } else {
        for (var i = 0; i < label_list.length; i++) {
            if (label_list[i] == checkID) {
                label_list.splice(i, 1)
                i--
            }
        }
    }

    console.log("after labels: " + label_list)
}

function makePost() {
    var labels = "";
    for (var i = 0; i < label_list.length; i++) {
        labels += label_list[i] + ";"
    }

    console.log("makePost: labels: " + labels)

    var e = document.getElementById("branchSelect");
    var b = e.options[e.selectedIndex].text;

    console.log("makePost: branch: " + b)

    post("/analyze", { branch: b, labels: labels })
}

function post(path, parameters) {
    var form = $('<form></form>');

    form.attr("method", "post");
    form.attr("action", path);

    $.each(parameters, function(key, value) {
        var field = $('<input></input>');

        field.attr("type", "hidden");
        field.attr("name", key);
        field.attr("value", value);

        console.log("adding: key=" + key + ", value=" + value)

        form.append(field);
    });

    $(document.body).append(form);
    form.submit();
}

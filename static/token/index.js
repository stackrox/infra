'use strict';

function handle(xhttp, success, failure) {
    xhttp.onreadystatechange = function() {
        if (this.readyState === 4){
            const obj = JSON.parse(xhttp.responseText);
            if (this.status === 200) {
                if (typeof success === 'function') {
                    return success(obj);
                }
                console.error("tried to call success function, but it was null");
            } else {
                if (typeof failure === 'function') {
                    return failure(obj);
                }
                console.error("tried to call failure function, but it was null");
            }
        }
    };
}

function generateToken(success, failure){
    var xhttp = new XMLHttpRequest();
    handle(xhttp, success, failure);
    xhttp.open("POST", "/v1/token", true);
    xhttp.setRequestHeader("Content-Type", "application/json;charset=UTF-8");
    xhttp.send();
}

function main() {
    const generate = document.getElementById("generate");
    const output = document.getElementById("output");

    generate.onclick = function(){
        generateToken(function (resp) {
            console.log(resp);
            var text = `# Run the following in a terminal to configure infractl for use:\nexport INFRA_TOKEN='${resp.Token}'`;
            output.innerText = text;
            output.style.display = "inline-block"
        }, function (resp) {
            console.error(resp);
        })
    };
}

(function(){
    window.onload = main;
})();

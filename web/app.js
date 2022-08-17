function postPass() {
    let pass = prompt("Please enter wallet password", "");
    if (pass != null) {
        var xhr = new XMLHttpRequest();
        xhr.open("POST", "/tdexconnect", true);
        xhr.setRequestHeader('password', pass);
        xhr.send()

        xhr.onload = function() {
            if (this.status === 200) {
                document.getElementById("url").innerHTML = this.responseText;
            }else {
                console.log(this.responseText);
                alert("Wrong password !!!");
            }
        }
    }
}
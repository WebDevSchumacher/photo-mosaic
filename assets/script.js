setupModals = function () {
    let registermodal = document.getElementById("register-modal");
    let loginmodal = document.getElementById("login-modal");
    let messagemodal = document.getElementById("message-modal");
    let registerbtn = document.getElementById("register-link");
    let loginbtn = document.getElementById("login-link");
    let registerclose = document.getElementById("register-close");
    let loginclose = document.getElementById("login-close");
    let messageclose = document.getElementById("message-close");

    registerbtn.addEventListener("click", function () {
        registermodal.style.display = "block";
    });
    loginbtn.addEventListener("click", function () {
        loginmodal.style.display = "block";
    });
    registerclose.addEventListener("click", function () {
        registermodal.style.display = "none";
    });
    loginclose.addEventListener("click", function () {
        loginmodal.style.display = "none";
    });
    messageclose.addEventListener("click", function () {
        messagemodal.style.display = "none";
    });
    window.addEventListener("click", function (event) {
        if (event.target === registermodal) {
            registermodal.style.display = "none";
        }
        if (event.target === loginmodal) {
            loginmodal.style.display = "none";
        }
        if (event.target === messagemodal) {
            messagemodal.style.display = "none";
        }
    });
    console.log("modals set up");
};
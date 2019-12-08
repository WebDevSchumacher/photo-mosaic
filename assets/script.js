setupNavModals = function () {
    let registermodal = document.getElementById("register-modal");
    let loginmodal = document.getElementById("login-modal");
    let messagemodal = document.getElementById("message-modal");
    let registerbtn = document.getElementById("register-link");
    let loginbtn = document.getElementById("login-link");
    let registerclose = document.getElementById("register-close");
    let loginclose = document.getElementById("login-close");
    let messageclose = document.getElementById("message-close");
    if (registerbtn !== null) {
        registerbtn.addEventListener("click", function () {
            registermodal.style.display = "block";
        });
    }
    if (loginbtn !== null) {
        loginbtn.addEventListener("click", function () {
            loginmodal.style.display = "block";
        });
    }
    if (registerclose !== null) {
        registerclose.addEventListener("click", function () {
            registermodal.style.display = "none";
        });
    }
    if (loginclose !== null) {
        loginclose.addEventListener("click", function () {
            loginmodal.style.display = "none";
        });
    }
    if (messageclose !== null) {
        messageclose.addEventListener("click", function () {
            messagemodal.style.display = "none";
        });
    }
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
};
setupNavHandlers = function () {
    let loginbtn = document.getElementById("login-button");
    let registerbtn = document.getElementById("register-button");
    let baseimageslink = document.getElementById("base-images-link");
    let loginXhr = new XMLHttpRequest();
    if (loginbtn !== null) {
        loginbtn.addEventListener("click", function () {
            loginXhr.open("POST", 'http://localhost:4242/login');
            let formdata = new FormData(document.getElementById("login-form"));
            loginXhr.send(formdata);
        });
    }
    loginXhr.addEventListener("load", function () {
        let response = JSON.parse(loginXhr.responseText);
        let modals = document.getElementsByClassName("modal");
        for (let i = 0; i < modals.length; i++) {
            modals[i].style.display = "none";
        }
        document.getElementById("message-text").innerText = response.Message;
        document.getElementById("message-modal").style.display = "block";
        if (response.Success && response.Path === "login") {
            window.location.reload();
        }
    });
    let registerXhr = new XMLHttpRequest();
    if (registerbtn !== null) {
        registerbtn.addEventListener("click", function () {
            registerXhr.open("POST", 'http://localhost:4242/register');
            let formdata = new FormData(document.getElementById("register-form"));
            registerXhr.send(formdata);
        });
    }
    registerXhr.addEventListener("load", function () {
        let response = JSON.parse(registerXhr.responseText);
        let modals = document.getElementsByClassName("modal");
        for (let i = 0; i < modals.length; i++) {
            modals[i].style.display = "none";
        }
        document.getElementById("message-text").innerText = response.Message;
        document.getElementById("message-modal").style.display = "block";
    });
    let baseImagesXhr = new XMLHttpRequest();
    if (baseimageslink !== null) {
        baseimageslink.addEventListener("click", function () {
            baseImagesXhr.open("GET", 'http://localhost:4242/base-images');
            baseImagesXhr.send();
        });
    }
    baseImagesXhr.addEventListener("load", function () {
        let response = JSON.parse(baseImagesXhr.responseText);
        document.getElementById("inner-content").innerHTML = response.Message;
        setupBaseListingControls();
        setupBaseListingItems();
    });
};
setupBaseListingItems = function () {
    let items = document.getElementsByClassName("listing-item");
    for (let i = 0; i < items.length; i++) {
        items[i].addEventListener("click", function (event) {
            // let loadBaseSetXhr = new XMLHttpRequest();
            // loadBaseSetXhr.open("GET", 'http://localhost:4242/base-images/get-set?setId='+event.target.id);
            // event.target.classList
            let active = document.getElementsByClassName("listing-item-active");
            if(active.length > 0){
                active[0].classList.remove("listing-item-active");
            }
            event.target.classList.add("listing-item-active")
        });
    }
};
setupBaseListingControls = function () {
    let newbasesetbtn = document.getElementById("new-base-set");
    // let editbaseset = document.getElementById("edit-base-set");
    // let deletebaseset = document.getElementById("delete-base-set");
    let newbasesetmodal = document.getElementById("new-base-set-modal");
    let newbasesetconfirm = document.getElementById("new-base-set-button");
    let newbasesetclose = document.getElementById("new-base-set-close");

    if (newbasesetbtn !== null) {
        newbasesetbtn.addEventListener("click", function () {
            newbasesetmodal.style.display = "block";
        });
    }
    let newBaseSetXhr = new XMLHttpRequest();
    if (newbasesetconfirm !== null) {
        newbasesetconfirm.addEventListener("click", function () {
            newBaseSetXhr.open("POST", 'http://localhost:4242/base-images/new-set');
            let formdata = new FormData(document.getElementById("new-base-set-form"));
            newBaseSetXhr.send(formdata);
        });
    }
    if (newbasesetclose !== null) {
        newbasesetclose.addEventListener("click", function () {
            newbasesetmodal.style.display = "none";
        });
    }
    newBaseSetXhr.addEventListener("load", function () {
        let response = JSON.parse(newBaseSetXhr.responseText);
        let modals = document.getElementsByClassName("modal");
        for (let i = 0; i < modals.length; i++) {
            modals[i].style.display = "none";
        }
        if (!response.Success) {
            document.getElementById("message-text").innerText = response.Message;
            document.getElementById("message-modal").style.display = "block";
        } else {
            let baseImagesXhr = new XMLHttpRequest();
            baseImagesXhr.open("GET", 'http://localhost:4242/base-images');
            baseImagesXhr.send();
            baseImagesXhr.addEventListener("load", function () {
                let response = JSON.parse(baseImagesXhr.responseText);
                document.getElementById("inner-content").innerHTML = response.Message;
                setupBaseListingControls();
                setupBaseListingItems();
            });
        }
    });
    window.addEventListener("click", function (event) {
        if (event.target === newbasesetmodal) {
            newbasesetmodal.style.display = "none";
        }
    });
};


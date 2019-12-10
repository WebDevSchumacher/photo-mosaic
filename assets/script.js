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
    let tilepoolslink = document.getElementById("tile-pools-link");
    let mosaiccollectionslink = document.getElementById("mosaic-collections-link");
    let loginXhr = new XMLHttpRequest();
    if (loginbtn !== null) {
        loginbtn.addEventListener("click", function () {
            loginXhr.open("POST", 'http://localhost:4242/picx/login');
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
            registerXhr.open("POST", 'http://localhost:4242/picx/register');
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
            baseImagesXhr.open("GET", 'http://localhost:4242/picx/base-images');
            baseImagesXhr.send();
        });
    }
    baseImagesXhr.addEventListener("load", function () {
        let response = JSON.parse(baseImagesXhr.responseText);
        document.getElementById("inner-content").innerHTML = response.Message;
        setupBaseListingControls();
        setupBaseListingItems();
    });
    let tilePoolsXhr = new XMLHttpRequest();
    if (tilepoolslink !== null) {
        tilepoolslink.addEventListener("click", function () {
            tilePoolsXhr.open("GET", 'http://localhost:4242/picx/tile-pools');
            tilePoolsXhr.send();
        });
    }
    tilePoolsXhr.addEventListener("load", function () {
        let response = JSON.parse(tilePoolsXhr.responseText);
        document.getElementById("inner-content").innerHTML = response.Message;
        setupTileListingControls();
        setupTileListingItems();
    });
    let mosaicCollectionsXhr = new XMLHttpRequest();
    if (mosaiccollectionslink !== null) {
        mosaiccollectionslink.addEventListener("click", function () {
            mosaicCollectionsXhr.open("GET", 'http://localhost:4242/picx/mosaic-collections');
            mosaicCollectionsXhr.send();
        });
    }
    mosaicCollectionsXhr.addEventListener("load", function () {
        let response = JSON.parse(mosaicCollectionsXhr.responseText);
        document.getElementById("inner-content").innerHTML = response.Message;
        setupMosaicListingControls();
        setupMosaicListingItems();
    });
};
setupBaseListingItems = function () {
    let items = document.getElementsByClassName("listing-item");
    for (let i = 0; i < items.length; i++) {
        items[i].addEventListener("click", function (event) {
            let active = document.getElementsByClassName("listing-item-active");
            if(active.length > 0){
                active[0].classList.remove("listing-item-active");
            }
            event.target.classList.add("listing-item-active");
            let loadBaseSetXhr = new XMLHttpRequest();
            loadBaseSetXhr.open("GET", 'http://localhost:4242/picx/base-images/get-set?setId='+event.target.id);
            loadBaseSetXhr.send();
            loadBaseSetXhr.addEventListener("load", function () {
                let response = JSON.parse(loadBaseSetXhr.responseText);
                document.getElementsByClassName("setbrowser-container")[0].innerHTML = response.Message;
            });
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
            newBaseSetXhr.open("POST", 'http://localhost:4242/picx/base-images/new-set');
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
            baseImagesXhr.open("GET", 'http://localhost:4242/picx/base-images');
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
uploadBaseSubmit = function (target) {
    let formdata = new FormData(target);
    let uploadXhr = new XMLHttpRequest();
    uploadXhr.open("POST", "/picx/base-images/upload");
    uploadXhr.send(formdata);
    uploadXhr.addEventListener("load", function () {
        document.getElementById(formdata.get("set-id").toString()).click();
    });
};
setupDetailsModal = function (id) {
    let modal = document.getElementById(id+"-modal");
    if(modal !== null){
        let close = document.getElementById(id+"-close");
        modal.style.display = "block";
        close.addEventListener("click", function () {
            modal.style.display = "none";
        });
        modal.addEventListener("click", function () {
            modal.style.display = "none";
        });
    }
};
setupTileListingItems = function () {
    let items = document.getElementsByClassName("listing-item");
    for (let i = 0; i < items.length; i++) {
        items[i].addEventListener("click", function (event) {
            let active = document.getElementsByClassName("listing-item-active");
            if(active.length > 0){
                active[0].classList.remove("listing-item-active");
            }
            event.target.classList.add("listing-item-active");
            let loadTilePoolXhr = new XMLHttpRequest();
            loadTilePoolXhr.open("GET", 'http://localhost:4242/picx/tile-pools/get-pool?poolId='+event.target.id);
            loadTilePoolXhr.send();
            loadTilePoolXhr.addEventListener("load", function () {
                let response = JSON.parse(loadTilePoolXhr.responseText);
                document.getElementsByClassName("poolbrowser-container")[0].innerHTML = response.Message;
            });
        });
    }
};
setupTileListingControls = function () {
    let newtilepoolbtn = document.getElementById("new-tile-pool");
    // let editbaseset = document.getElementById("edit-base-set");
    // let deletebaseset = document.getElementById("delete-base-set");
    let newtilepoolmodal = document.getElementById("new-tile-pool-modal");
    let newtilepoolconfirm = document.getElementById("new-tile-pool-button");
    let newtilepoolclose = document.getElementById("new-tile-pool-close");

    if (newtilepoolbtn !== null) {
        newtilepoolbtn.addEventListener("click", function () {
            newtilepoolmodal.style.display = "block";
        });
    }
    let newTilePoolXhr = new XMLHttpRequest();
    if (newtilepoolconfirm !== null) {
        newtilepoolconfirm.addEventListener("click", function () {
            newTilePoolXhr.open("POST", 'http://localhost:4242/picx/tile-pools/new-pool');
            let formdata = new FormData(document.getElementById("new-tile-pool-form"));
            newTilePoolXhr.send(formdata);
        });
    }
    if (newtilepoolclose !== null) {
        newtilepoolclose.addEventListener("click", function () {
            newtilepoolmodal.style.display = "none";
        });
    }
    newTilePoolXhr.addEventListener("load", function () {
        let response = JSON.parse(newTilePoolXhr.responseText);
        let modals = document.getElementsByClassName("modal");
        for (let i = 0; i < modals.length; i++) {
            modals[i].style.display = "none";
        }
        if (!response.Success) {
            document.getElementById("message-text").innerText = response.Message;
            document.getElementById("message-modal").style.display = "block";
        } else {
            let tilePoolsXhr = new XMLHttpRequest();
            tilePoolsXhr.open("GET", 'http://localhost:4242/picx/tile-pools');
            tilePoolsXhr.send();
            tilePoolsXhr.addEventListener("load", function () {
                let response = JSON.parse(tilePoolsXhr.responseText);
                document.getElementById("inner-content").innerHTML = response.Message;
                setupTileListingControls();
                setupTileListingItems();
            });
        }
    });
    window.addEventListener("click", function (event) {
        if (event.target === newtilepoolmodal) {
            newtilepoolmodal.style.display = "none";
        }
    });
};
uploadTileSubmit = function (target) {
    let formdata = new FormData(target);
    let uploadXhr = new XMLHttpRequest();
    uploadXhr.open("POST", "/picx/tile-pools/upload");
    uploadXhr.send(formdata);
    uploadXhr.addEventListener("load", function () {
        document.getElementById(formdata.get("pool-id").toString()).click();
    });
};

setupMosaicListingItems = function () {
    let items = document.getElementsByClassName("listing-item");
    for (let i = 0; i < items.length; i++) {
        items[i].addEventListener("click", function (event) {
            let active = document.getElementsByClassName("listing-item-active");
            if(active.length > 0){
                active[0].classList.remove("listing-item-active");
            }
            event.target.classList.add("listing-item-active");
            let loadMosaicCollectionXhr = new XMLHttpRequest();
            loadMosaicCollectionXhr.open("GET", 'http://localhost:4242/picx/mosaic-collections/get-collection?collectionId='+event.target.id);
            loadMosaicCollectionXhr.send();
            loadMosaicCollectionXhr.addEventListener("load", function () {
                let response = JSON.parse(loadMosaicCollectionXhr.responseText);
                document.getElementsByClassName("collectionbrowser-container")[0].innerHTML = response.Message;
            });
        });
    }
};
setupMosaicListingControls = function () {
    let newmosaiccollectionbtn = document.getElementById("new-mosaic-collection");
    // let editbaseset = document.getElementById("edit-base-set");
    // let deletebaseset = document.getElementById("delete-base-set");
    let newmosaiccollectionmodal = document.getElementById("new-mosaic-collection-modal");
    let newmosaiccollectionconfirm = document.getElementById("new-mosaic-collection-button");
    let newmosaiccollectionclose = document.getElementById("new-mosaic-collection-close");

    if (newmosaiccollectionbtn !== null) {
        newmosaiccollectionbtn.addEventListener("click", function () {
            newmosaiccollectionmodal.style.display = "block";
        });
    }
    let newMosaicCollectionXhr = new XMLHttpRequest();
    if (newmosaiccollectionconfirm !== null) {
        newmosaiccollectionconfirm.addEventListener("click", function () {
            newMosaicCollectionXhr.open("POST", 'http://localhost:4242/picx/mosaic-collections/new-collection');
            let formdata = new FormData(document.getElementById("new-mosaic-collection-form"));
            newMosaicCollectionXhr.send(formdata);
        });
    }
    if (newmosaiccollectionclose !== null) {
        newmosaiccollectionclose.addEventListener("click", function () {
            newmosaiccollectionmodal.style.display = "none";
        });
    }
    newMosaicCollectionXhr.addEventListener("load", function () {
        let response = JSON.parse(newMosaicCollectionXhr.responseText);
        let modals = document.getElementsByClassName("modal");
        for (let i = 0; i < modals.length; i++) {
            modals[i].style.display = "none";
        }
        if (!response.Success) {
            document.getElementById("message-text").innerText = response.Message;
            document.getElementById("message-modal").style.display = "block";
        } else {
            let tilePoolsXhr = new XMLHttpRequest();
            tilePoolsXhr.open("GET", 'http://localhost:4242/picx/mosaic-collections');
            tilePoolsXhr.send();
            tilePoolsXhr.addEventListener("load", function () {
                let response = JSON.parse(tilePoolsXhr.responseText);
                document.getElementById("inner-content").innerHTML = response.Message;
                setupMosaicListingControls();
                setupMosaicListingItems();
            });
        }
    });
    window.addEventListener("click", function (event) {
        if (event.target === newmosaiccollectionmodal) {
            newmosaiccollectionmodal.style.display = "none";
        }
    });
};
//Connect form
var connectForm = document.getElementById('connectForm');
var ipField = document.getElementById("ipField");
var portField = document.getElementById("portField");
var okConnect = document.getElementById('connectOK');
var failConnect = document.getElementById('connectFail');
var waitConnect = document.getElementById('connectWait');
var nodeID = document.getElementById('nodeID')
var connectButton = document.getElementById('connectButton');
var quitButton = document.getElementById('quitButton');

//Upload form
var uploadForm = document.getElementById('storeFile');
var uploadFileSelect = document.getElementById('fileToUpload');
var uploadFileButton = document.getElementById('uploadButton');
var oKIconUpload = document.getElementById('uploadOK');
var failIconUpload = document.getElementById('uploadFail');
var waitIconUpload = document.getElementById('uploadWait');

//Download form
var downloadForm = document.getElementById('downloadFile');
var searchToDownload = document.getElementById('fileToDownload');
var searchToDownButton = document.getElementById('searchToDownload');
var downloadFileButton = document.getElementById('downloadButton');
var waitIconDownload = document.getElementById('downloadSearchRefresh');
var okIconDownload = document.getElementById('downloadSearchOK');
var failIconDownload = document.getElementById('downloadSearchFail');

//Update form
var updtateForm = document.getElementById('updateFile');
var searchToUpdate = document.getElementById('fileToUpdate');
var searchToUpdateButton = document.getElementById('searchToUpdate');
var updateFileButton = document.getElementById('updateButton');
var waitIconUpdate = document.getElementById('updateSearchRefresh');
var okIconUpdate = document.getElementById('updateSearchOK');
var failIconUpdate = document.getElementById('updateSearchFail');

//Delete form
var deleteForm = document.getElementById('deleteFile');
var searchToDelete = document.getElementById('fileToDelete');
var searchToDeleteButton = document.getElementById('searchToDelete');
var deleteFileButton = document.getElementById('deleteButton');
var waitIconDelete = document.getElementById('deleteSearchRefresh');
var okIconDelete = document.getElementById('deleteSearchOK');
var failIconDelete = document.getElementById('deleteSearchFail');

function getStatus(){
    var xhr = new XMLHttpRequest();
    xhr.open('GET', '/status', true);
    xhr.onload = function() {
        if(xhr.status === 200) {
            var res = xhr.responseText.split("/");
            nodeID.innerHTML = "Node ID : " + res[0];
            if(res[1]==="true")
            {
                connectButton.style.display = 'none'
                okConnect.style.display = 'inline'
                quitButton.style.display = 'inline'
            } else{
                quitButton.style.display = 'none'
            }
        } else {
            alert("Response fail");
        }
    }
    xhr.send();
}

connectForm.onsubmit = function(event){
    return false; //Prevent enter key pressed
}

function connect(){
    refreshUI("connect");
    waitConnect.style.display = 'inline';

    var xhr = new XMLHttpRequest();
    xhr.open('GET', '/join?ip='+ipField.value+'&port='+portField.value, true);
    xhr.onload = function(){
        waitConnect.style.display = 'none';
        if (xhr.status === 200) {
            if(xhr.responseText==="true"){
                okConnect.style.display = 'inline';
                connectButton.style.display = 'none';
                quitButton.style.display = 'inline';
            } else if (xhr.responseText === "false"){
                failConnect.style.display = 'inline';
            }
        } else {
           alert("Connection Error");
        }
    };
    xhr.send()
}

function leaveCircle(){
    var xhr = new XMLHttpRequest();
    xhr.open('GET', '/leave', true);
    xhr.send();
}

uploadForm.onsubmit = function(event) {
    event.preventDefault();
    refreshUI("upload");
    waitIconUpload.style.display = 'inline';
    uploadFileButton.innerHTML = 'Uploading...'
    var files = uploadFileSelect.files;
    var file = files[0];
    var formData = new FormData();

    formData.append('file', file, file.name);
    var xhr = new XMLHttpRequest();
    xhr.open('POST', '/store', true);

    xhr.onload = function(){
        refreshUI("upload");
        if (xhr.status === 200) {
            // File(s) uploaded.
            if(xhr.responseText === "FileExists"){
                alert("File already exists, use 'update' instead.")
                failIconUpload.style.display = 'inline';
            }
            uploadFileButton.innerHTML = 'Upload file';
            resetForms();
            if(xhr.responseText ==="OK")
            oKIconUpload.style.display = 'inline';
        } else {
            infirmUpload.style.display = 'inline';
            alert('Upload Failed');
        }
    };
    xhr.send(formData);
}

downloadForm.onsubmit = function(event){
    return false; //Prevent enter key pressed
}

function downloadFile(){
    downloadFileButton.innerHTML = 'Downloading...';
    var formData = new FormData();

    formData.append('value', searchToDownload.value);
    var xhr = new XMLHttpRequest();
    xhr.open('GET', '/download?value='+searchToDownload.value, true);
    xhr.responseType = "blob";

    xhr.onload = function(){
        if (xhr.status === 200) {
            refreshUI("download")
            // File(s) downloaded.
            downloadFileButton.innerHTML = 'Download file';
            download(xhr.response , searchToDownload.value, "yourFile");
            resetForms();
            
            //alert(xhr.response);
        } else {
            alert('Download Failed');
        }
    };
    xhr.send(formData);
}

updtateForm.onsubmit = function(event){
    return false; //Prevent enter key pressed
}

function update(){
    updateFileButton.innerHTML = 'Updating...';
    var files = searchToUpdate.files;
    var file = files[0];
    var formData = new FormData();

    formData.append('file', file, file.name);
    var xhr = new XMLHttpRequest();
    xhr.open('POST', '/update', true);

    xhr.onload = function(){
        if(xhr.status=== 200)
        {
            refreshUI("update")
            updateFileButton.innerHTML = "Update File";
            resetForms();
            if(xhr.responseText==="OK"){
                okIconUpdate.style.display = 'inline';
            }
            if(xhr.responseText==="NotFileExists"){
                failIconUpdate.style.display = 'inline';
                alert("File does not exists : use 'upload' instead.");
            }
        } else {
            failIconUpdate.style.display = 'inline';
            alert("Update failed");
        }
    };
    xhr.send(formData);
}

deleteForm.onsubmit = function(event){
    return false; //Prevent enter key pressed
}

function deleteFile(){
    deleteFileButton.innerHTML = "Deleting file...";
    var formData = new FormData();

    formData.append('value', searchToDelete.value);
    var xhr = new XMLHttpRequest();
    xhr.open('DELETE', '/delete?value='+searchToDelete.value, true);

    xhr.onload = function(){
        if (xhr.status === 200){
            refreshUI("delete")
            deleteFileButton.innerHTML = 'Delete file';
            resetForms();

        } else {
            alert('Delete failed!');
        }
    };
    xhr.send(formData);
}


// searchToDownload.on("change", test())
//searchToDownload.keyput(test())
function refreshUI(field){
    if (field === "download"){
        downloadFileButton.style.display = 'none';
        okIconDownload.style.display = 'none';
        failIconDownload.style.display = 'none';
    } else if (field === "update"){
        updateFileButton.style.display = 'none';
        okIconUpdate.style.display = 'none';
        failIconUpdate.style.display = 'none';
    } else if (field === "delete"){
        deleteFileButton.style.display = 'none';
        okIconDelete.style.display = 'none';
        failIconDelete.style.display = 'none';
    } else if (field==="upload"){
        oKIconUpload.style.display = 'none';
        waitIconUpload.style.display = 'none';
        failIconUpload.style.display = 'none';
    } else if (field==="connect"){
        okConnect.style.display = 'none';
        failConnect.style.display = 'none';
        waitConnect.style.display = 'none';
    }
}

function search(toSearch, field){
    var button;
    var wait;
    var ok;
    var fail;
    var submit;
    if (field === "download")
    {
        button = searchToDownButton;
        wait = waitIconDownload;
        ok = okIconDownload;
        fail = failIconDownload;
        submit = downloadButton;
    } else if (field === "update")
    {
        button = searchToUpdateButton;
        wait = waitIconUpdate;
        ok = okIconUpdate;
        fail = failIconUpdate;
        submit = updateFileButton;
    } else if (field === "delete")
    {
        button = searchToDeleteButton;
        wait = waitIconDelete;
        ok = okIconDelete;
        fail = failIconDelete;
        submit = deleteFileButton;
    }
    button.innerHTML = "Searching..."
    wait.style.display = 'inline'
    var xhr = new XMLHttpRequest();
    xhr.open('GET', '/search?value='+toSearch, true);
    xhr.onload = function(){
        if (xhr.status === 200) {
            // File(s) uploaded.
            button.innerHTML = 'Search';
            wait.style.display='none';
            if (xhr.responseText === "false")
            {
                ok.style.display = 'inline';
                submit.style.display = 'inline';
            } else
            {
                fail.style.display = 'inline';
            }

        } else {
            alert('Search Failed');
        }
    };
    xhr.send("value:lol");
}



function resetForms(){
    uploadForm.reset();
    downloadForm.reset();
    updtateForm.reset();
    deleteForm.reset();
}
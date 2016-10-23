//Upload form
var uploadForm = document.getElementById('storeFile');
var uploadFileSelect = document.getElementById('fileToUpload');
var uploadFileButton = document.getElementById('uploadButton');

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


uploadForm.onsubmit = function(event) {
    event.preventDefault();
    uploadFileButton.innerHTML = 'Uploading...'
    var files = uploadFileSelect.files;
    var file = files[0];
    var formData = new FormData();
    var confirmUpload = document.getElementById('uploadOK');
    var infirmUpload = document.getElementById('uploadFail');

    formData.append('file', file, file.name);
    var xhr = new XMLHttpRequest();
    xhr.open('POST', '/store', true);

    xhr.onload = function(){
        if (xhr.status === 200) {
            // File(s) uploaded.
            uploadFileButton.innerHTML = 'Upload file';
            resetForms();
            confirmUpload.style.display = 'inline';
        } else {
            infirmUpload.style.display = 'inline';
            alert('Upload Failed');
        }
    };
    xhr.send(formData);
}

function download(){
    downloadFileButton.innerHTML = 'Downloading...';
    var formData = new FormData();

    formData.append('value', searchToDownload.value);
    var xhr = new XMLHttpRequest();
    xhr.open('POST', '/download', true);

    xhr.onload = function(){
        if (xhr.status === 200) {
            // File(s) downloaded.
            downloadFileButton.innerHTML = 'Download file';
            resetForms();
        } else {
            alert('Download Failed');
        }
    };
    xhr.send(formData);
}

function update(){
    updateFileButton.innerHTML = 'Updating...';
    var files = searchToUpdate.files;
    var file = files[0];
    var formData = new FormData();

    formData.append('file', file, file.name);
    var xhr = new XMLHttpRequest();
    xhr.open('PUT', '/update', true);

    xhr.onload = function(){
        if(xhr.status=== 200)
        {
            updateFileButton.innerHTML = "Update File";
            resetForms();
        } else {
            alert("Update failed");
        }
    };
    xhr.send(formData);
}

function deleteFile(){
    deleteFileButton.innerHTML = "Deleting file...";
    var formData = new FormData();

    formData.append('value', searchToDelete.value);
    var xhr = new XMLHttpRequest();
    xhr.open('DELETE', '/delete', true);

    xhr.onload = function(){
        if (xhr.status === 200){
            deleteFileButton.innerHTML = 'Delete file';
            resetForms();
        } else {
            alert('Update failed!');
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
        failIconDownload.display = 'none';
    } else if (field === "update"){
        updateFileButton.style.display = 'none';
        okIconUpdate.style.display = 'none';
        failIconUpdate.style.display = 'none';
    } else if (field === "delete"){
        deleteFileButton.style.display = 'none';
        okIconDelete.style.display = 'none';
        failIconDelete.style.display = 'none';
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
    var formData = new FormData();
    formData.append("value", toSearch);
    var xhr = new XMLHttpRequest();
    xhr.open('GET', '/search', true);
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
    xhr.send(formData);
}



function resetForms(){
    uploadForm.reset();
    downloadForm.reset();
    updtateForm.reset();
    deleteForm.reset();
}
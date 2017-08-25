"use strict";

let controls = {
  btns: document.querySelectorAll('#divControls button'),
  btnsPlaylist: document.querySelectorAll('#divControlPlaylist'),
  btnStart: document.querySelector('#btnStart'),
  spCurrent: document.querySelector('#spCurrent'),
  tblPlaylist: document.querySelector('#tblPlaylist')
}


let playlist = {
  items: Array.from(controls.tblPlaylist.querySelectorAll('td')).map(el => el.textContent),
  selected: null,
  playPause: e => console.log(e)
}

controls.tblPlaylist.addEventListener('click', plSelect);
controls.btns.forEach(btn => btn.addEventListener('click', sendCommand));
controls.btnsPlaylist.forEach(btn => btn.addEventListener('click', callMethod));
controls.btnStart.addEventListener('click', startItem);

function plSelect(e) {
  if (playlist.selected != null) {
    playlist.selected.classList.remove('selected');
  }
  playlist.selected = e.target;
  e.target.classList.add('selected');
}

function callMethod(e) {
  let reqBody = {
    component: "player",
    method: e.target.dataset["method"]
  };

  callApi(reqBody)
}

function startItem(e) {
  let reqBody = {
    component: "player",
    method: "start",
    arguments: {
      path: playlist.selected.textContent
    }
  };

  callApi(reqBody)
    .then(json => {
      if (json.success) {
        spCurrent.textContent = json.message
      }
    });
}

function sendCommand(e) {
  let command = e.target.dataset["command"];
  let reqBody = {
    component: "player",
    method: "sendCommand",
    arguments: {
      command: command
    }
  };

  callApi(reqBody);
}

function callApi(reqBody) {
  let myHeaders = new Headers();
  myHeaders.append('Content-Type', 'application/json');

  let myInit = {
    method: "POST",
    headers: myHeaders,
    body: JSON.stringify(reqBody)
  }

  return fetch(`${window.location.origin}/api`, myInit)
    .then(res => res.json())
    .then(json => {
      console.log(json);
      return json;
    })
    .catch(err => console.error(err));
}
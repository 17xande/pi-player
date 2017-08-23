"use strict";

const btns = document.querySelectorAll('#divControls button');
const btnStart = document.querySelector('#btnStart');
const spPlaying = document.querySelector('#spPlaying');
const ulFiles = document.querySelector('#ulPlaylist');

let playlist = {
  items: Array.from(document.querySelectorAll('#ulPlaylist li')).map(el => el.textContent),
  selected: null,
  playPause: e => console.log(e)
}

ulPlaylist.addEventListener('click', plSelect);
btns.forEach(btn => btn.addEventListener('click', sendCommand));
btnStart.addEventListener('click', startItem);

function plSelect(e) {
  if (playlist.selected != null) {
    playlist.selected.classList.remove('selected');
  }
  playlist.selected = e.target;
  e.target.classList.add('selected');
}

function startItem(e) {
  let reqBody = {
    component: "player",
    method: "start",
    arguments: {
      path: playlist.selected.textContent
    }
  };

  callApi(reqBody);
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

  callApi(reqBody)
}

function callApi(reqBody) {
  let myHeaders = new Headers();
  myHeaders.append('Content-Type', 'application/json');

  let myInit = {
    method: "POST",
    headers: myHeaders,
    body: JSON.stringify(reqBody)
  }

  fetch(`${window.location.origin}/api`, myInit)
    .then(res => res.json())
    .then(json => console.log(json))
    .catch(err => console.error(err));
}
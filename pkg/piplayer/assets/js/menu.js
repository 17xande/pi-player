"use strict";

run();

function run() {
  if (!window["WebSocket"]) {
    console.error("This page requires WebSocket support. Please use a WebSocket enabled service.");
    return;
  }

  // Ignore all keyboard input on the Pi browser.
  document.addEventListener("keydown", e => {
    e.preventDefault();
  });

  const menuItemSelector = '.item';
  const divContainer = document.querySelector('#container');
  const ulPlaylist = document.querySelector('#ulPlaylist');
  const vidMedia = document.querySelector('#vidMedia');
  const audMusic = document.querySelector('#audMusic');
  const arrItems = Array.from(document.querySelectorAll(menuItemSelector));

  let conn = null;

  if (arrItems.length <= 0) {
    // No items in the menu. Nothing to do here.
    console.warn("No items in the playlist, so then not much to do here?")
    return;
  }

  wsConnect();

  function wsConnect() {
    conn = new WebSocket('ws://' + document.location.host + '/ws');

    // If connection is not established, try again after 2 seconds.
    let to = setTimeout(() => {
      if (conn.readyState != 1) {
        console.warn("Connection attempt unsuccessfull, trying again...");
        wsConnect();
      } else {
        console.log("Re-connection successful.");
      }
    }, 2000);
  }
  
  conn.addEventListener('open', e => {
    console.log("Connection Opened.");
  });
  
  conn.addEventListener('error', e => {
    console.log("Error in the websocket connection:\n", err);
  });

  conn.addEventListener('close', e => {
    console.log("Connection closed.\nTrying to reconnect...");

    let to = setTimeout(() => wsConnect(), 2000);
  });

  conn.addEventListener('message', e => {
    let msg = JSON.parse(e.data)

    console.log(msg);

    switch (msg.message) {
      case 'KEY_UP':
      case 'KEY_DOWN':
        remoteArrowPress(e, msg);
        break;
      case 'KEY_ENTER':
        remoteEnterPress(e, msg);
        break;
      case 'KEY_HOME':
        remoteHomePress(e, msg);
        break;
      default:
        console.log("Unsupported message received: ", e.data);
        break;
    }
  });

  function remoteArrowPress(e, msg) {
    let selectedItem = document.querySelector(menuItemSelector + ':focus');
    if (selectedItem == null) {
      // No item is selected, focus on first item.
      arrItems[0].focus();
      return;
    }

    let i = arrItems.indexOf(selectedItem);
    if (i < 0) {
      console.error("Element not in initial array of elements?\nFocusing on first item.")
      arrItems[0].focus();
      return;
    }

    let diff = msg.message == 'KEY_UP' ? -1 : 1;

    if (msg.message == 'KEY_UP' && i <= 0) {
      i = arrItems.length;
    } else if (msg.message == 'KEY_DOWN' && i >= arrItems.length - 1) {
      i = -1;
    }

    arrItems[i + diff].focus();
  }

  function remoteEnterPress(e, msg) {
    let selectedItem = document.querySelector(menuItemSelector + ':focus');

    if (selectedItem == null) {
      // No item selected, focus on first item again.
      arrItems[0].focus();
      return;
    }

    // let reqBody = {
    //   component: 'player',
    //   method: 'start',
    //   arguments: {
    //     path: selectedItem.textContent
    //   }
    // };
    // conn.send(JSON.stringify(reqBody));
    startItem(selectedItem);
  }

  function getItems() {
    let reqBody = {
      component: 'playlist',
      method: 'getItems'
    }

    conn.send(JSON.stringify(reqBody));
  }

  function remoteHomePress(e, msg) {
    // If the menu is hidden, show it.
    if (ulPlaylist.style.visibility !== 'visible') {
      ulPlaylist.style.visibility = 'visible';
      // arrItems[0].focus();
      return;
    }

    // If the meny is showing, hide it.
    if (ulPlaylist.style.visibility === 'visible') {
      ulPlaylist.style.visibility = 'hidden';
    }
  }

  function startItem(el) {
    let n = el.textContent;
    let ext = n.slice(n.lastIndexOf('.'));
    ulPlaylist.style.visibility = 'hidden';

    switch (ext) {
      case '.mp4':
        vidMedia.src = `/content/${n}`;
        vidMedia.style.visibility = 'visible';
        // Blackout the background.
        divContainer.style.backgroundImage = null;
        vidMedia.play();
        break;
      case '.jpg':
      case '.jpeg':
      case '.png':
        // Stop video if playing.
        if (!vidMedia.paused) {
          vidMedia.pause();
          vidMedia.style.visibility = 'hidden';
        }
        // Change background image.
        divContainer.style.backgroundImage = `url("/content/${n}")`;
      break;
      default:
        console.log("File type not supported: ", ext);
        break;
    }
  }
}
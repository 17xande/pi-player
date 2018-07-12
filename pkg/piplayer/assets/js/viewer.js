"use strict";

class Viewer {
  menuItemSelector = '.item';
  conn = null;
  arrItems = null;
  divContainer = document.querySelector('#container');
  ulPlaylist = document.querySelector('#ulPlaylist');
  vidMedia = document.querySelector('#vidMedia');
  audMusic = document.querySelector('#audMusic');
  playlist = {
    current: null,
    items: []
  };

  constructor() {
    if (!window["WebSocket"]) {
      console.error("This page requires WebSocket support. Please use a WebSocket enabled service.");
      return;
    }
  
    // Ignore all keyboard input on the Pi browser.
    // document.addEventListener("keydown", e => {
    //   e.preventDefault();
    // });
  
    this.arrItems = Array.from(document.querySelectorAll(this.menuItemSelector));
  
    if (this.arrItems.length <= 0) {
      // No items in the menu. Nothing to do here.
      console.warn("No items in the playlist, so then not much to do here?")
      return;
    }

    this.getItems().then(res => {
      this.startItem(0);
    });
    this.wsConnect();
  }

  wsConnect() {
    let u = 'ws://' + document.location.host + '/ws';
    this.conn = new WebSocket(u);

    this.conn.addEventListener('open', e => {
      console.log("Connection Opened.");
    });
    
    this.conn.addEventListener('error', e => {
      console.log("Error in the websocket connection:\n", e);
    });
  
    this.conn.addEventListener('close', e => {
      console.log("Connection closed.\nTrying to reconnect...");
  
      let to = setTimeout(() => this.wsConnect(), 2000);
    });

    this.conn.addEventListener('message', this.socketMessage.bind(this));
  }

  callApi(reqBody) {
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

  socketMessage(e) {
    let msg = JSON.parse(e.data)

    console.log(msg);

    switch (msg.message) {
      case 'KEY_UP':
      case 'KEY_DOWN':
        this.remoteArrowPress(e, msg);
        break;
      case 'KEY_LEFT':
        this.remoteArrowLeftPress(e);
        break;
      case 'KEY_RIGHT':
        this.remoteArrowRightPress(e);
        break;
      case 'KEY_ENTER':
        this.remoteEnterPress(e);
        break;
      case 'KEY_CONTEXT_MENU':
        this.remoteHomePress(e);
        break;
      case 'KEY_PLAYPAUSE':
        this.remotePlayPress(e);
        break;
      case 'KEY_STOP':
        this.remoteStopPress(e);
      case 'KEY_FASTFORWARD':
        this.remoteSeek(e, 15);
        break;
      case 'KEY_REWIND':
        this.remoteSeek(e, -15);
        break;
      default:
        console.log("Unsupported message received: ", e.data);
        break;
    }
  }

  getItems() {
    let reqBody = {
      component: 'playlist',
      method: 'getItems'
    }

    return this.callApi(reqBody)
      .then(res => {
        if (!res || !res.success) {
          console.error(res);
          return;
        }
        this.playlist.items = res.message;
        return res;
      });
  }

  remoteArrowPress(e, msg) {
    let selectedItem = document.querySelector(this.menuItemSelector + ':focus');
    if (selectedItem == null) {
      // No item is selected, focus on first item.
      this.arrItems[0].focus();
      return;
    }

    let i = this.arrItems.indexOf(selectedItem);
    if (i < 0) {
      console.error("Element not in initial array of elements?\nFocusing on first item.")
      this.arrItems[0].focus();
      return;
    }

    let diff = msg.message == 'KEY_UP' ? -1 : 1;

    if (msg.message == 'KEY_UP' && i <= 0) {
      i = this.arrItems.length;
    } else if (msg.message == 'KEY_DOWN' && i >= this.arrItems.length - 1) {
      i = -1;
    }

    this.arrItems[i + diff].focus();
  }

  remoteArrowLeftPress(e) {
    if (this.playlist.current == 0) {
      this.startItem(this.playlist.items.length - 1);
      return
    }

      this.startItem(this.playlist.current - 1);
  }

  remoteArrowRightPress(e) {
    if (this.playlist.current >= this.playlist.items.length - 1) {
      this.startItem(0);
      return;
    }

    this.startItem(this.playlist.current + 1);
  }

  remoteEnterPress(e) {
    let selectedItem = document.querySelector(this.menuItemSelector + ':focus');

    if (selectedItem == null) {
      // No item selected, focus on first item again.
      this.arrItems[0].focus();
      return;
    }

    let i = selectedItem.dataset.index;
    this.startItem(i);
  }

  remoteHomePress(e) {
    // If the menu is hidden, show it.
    if (this.ulPlaylist.style.visibility !== 'visible') {
      this.ulPlaylist.style.visibility = 'visible';
      this.arrItems[this.playlist.current].focus();
      return;
    }

    // If the meny is showing, hide it.
    if (this.ulPlaylist.style.visibility === 'visible') {
      this.ulPlaylist.style.visibility = 'hidden';
    }
  }

  remotePlayPress(e) {
    if (this.vidMedia.paused) {
      this.vidMedia.play();
    } else {
      this.vidMedia.pause();
    }
  }

  remoteStopPress(e) {
    this.vidMedia.pause();
    this.videMedia.currentTime = 0;
  }

  remoteSeek(e, msg) {
    this.vidMedia.currentTime += msg;
  }

  startItem(index) {
    if (index <= -1) {
      console.error("Cannot play item at negative index.");
      return;
    }

    let name = this.playlist.items[index].Visual;
    let ext = name.slice(name.lastIndexOf('.'));
    this.ulPlaylist.style.visibility = 'hidden';

    switch (ext) {
      case '.mp4':
        this.vidMedia.src = `/content/${name}`;
        this.vidMedia.style.visibility = 'visible';
        // Blackout the background.
        this.divContainer.style.backgroundImage = null;
        this.vidMedia.play();
        break;
      case '.jpg':
      case '.jpeg':
      case '.png':
        // Stop video if playing.
        if (!this.vidMedia.paused) {
          this.vidMedia.pause();
          this.vidMedia.style.visibility = 'hidden';
        }
        // Change background image.
        this.divContainer.style.backgroundImage = `url("/content/${name}")`;
      break;
      default:
        console.log("File type not supported: ", ext);
        return;
        break;
    }
    this.playlist.current = index;

    let reqBody = {
      component: "playlist",
      method: "setCurrent",
      arguments: {index: index}
    };

    this.callApi(reqBody).then(res => {
      if (!res || !res.success) {
        console.error("Cound't set the current item through the API.");
      }
    });
  }
}

let viewer = new Viewer();
class Control {
  constructor() {
    // make these constants in a module
    this.btns = document.querySelectorAll('#divControlsPlayer button');
    this.btnsPlaylist = document.querySelectorAll('#divControlPlaylist');
    this.btnStart = document.querySelector('#btnStart');
    this.spCurrent = document.querySelector('#spCurrent');
    this.tblPlaylist = document.querySelector('#tblPlaylist');
    this.wsPath = "/ws/control";

    this.conn = null;
    this.playlist = {
      current: null,
      selected: null,
      items: []
    };

    if (!window["WebSocket"]) {
      console.error("This page requires WebSocket support. Please use a WebSocket enabled service.");
      return;
    }

    this.getItems().then(res => {
      console.log("loaded playlist from server");
    })

    this.wsConnect();

    this.tblPlaylist.addEventListener('click', this.plSelect.bind(this));
    this.btns.forEach(btn => btn.addEventListener('click', this.sendCommand.bind(this)));
    this.btnsPlaylist.forEach(btn => btn.addEventListener('click', this.callMethod.bind(this)));
    this.btnStart.addEventListener('click', this.startItem.bind(this));
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

  wsConnect() {
    let u = 'ws://' + document.location.host + this.wsPath;
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

  socketMessage(e) {
    let msg = JSON.parse(e.data);
    console.log(msg);

    switch (msg.event) {
      case "setCurrent":
        this.setCurrent(parseInt(msg.message))
        break;
      default:
      console.log("Unsupported message received: ", e.data);
    }
  }

  plSelect(e) {
    if (this.playlist.selected != null) {
      this.playlist.selected.classList.remove('selected');
    }
    this.playlist.selected = e.target.closest('tr');
    this.playlist.selected.classList.add('selected');
  }

  setCurrent(index) {
    this.playlist.current = index;
    this.spCurrent.textContent = this.playlist.items[index].Visual;
    let el = this.tblPlaylist.querySelector(`tr[data-index="${index}"]`);
    this.plSelect({target: el});
  }
  
  callMethod(e) {
    let btn = e.target.closest('button');
  
    let reqBody = {
      component: "player",
      method: btn.dataset["method"]
    };
  
    this.callApi(reqBody)
      .then(this.videoCallback.bind(this));
  }
  
  startItem(e) {
    let s = this.playlist.selected;
    let itemName = s.querySelector('td.item-name').textContent;
    let reqBody = {
      component: "player",
      method: "start",
      arguments: {
        path: itemName,
        index: s.dataset.index
      }
    };
  
    this.callApi(reqBody).then(this.videoCallback.bind(this));
  }
  
  videoCallback(json) {
    if (json.success) {
      console.log("instruction sent successfully. awaiting confirmation in socket.");
    }
  }
  
  sendCommand(e) {
    let btn = e.target.closest('button');
    
    let reqBody = {
      component: "player",
      method: "sendCommand",
      arguments: {
        command: btn.dataset["command"]
      }
    };
  
    this.callApi(reqBody);
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
}

let control = new Control();
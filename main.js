const {
  app,
  BrowserWindow,
  shell,
  ipcMain,
  dialog
} = require('electron');

const fs = require('fs');
const http = require('http');
const url = require('url');
const path = require('path');

const {google} = require('googleapis');
const OAuth2 = google.auth.OAuth2;
const youtube = google.youtube('v3');

const SCOPES = [
  'https://www.googleapis.com/auth/youtube.readonly',
  'https://www.googleapis.com/auth/youtube.channel-memberships.creator'
];

const CLIENT_ID = process.env.GOOGLE_CLIENT_ID;
const CLIENT_SECRET = process.env.GOOGLE_CLIENT_SECRET;
const PROJECT_ID = "tukija";
const TOKEN_PATH = path.join(app.getPath('userData'), '.token.json');
const CONFIG_PATH = path.join(app.getPath('userData'), 'config.json');

let config = {
  width: 800,
  height: 600,
  state: {}
}

const main = () => {
  const win = new BrowserWindow({
    width: config.width,
    height: config.height,
    webPreferences: {
      nodeIntegration: true
    }
  });

  win.setMenu(null);

  let isAuthenticated = false;
  let oauth2Client = new OAuth2(CLIENT_ID, CLIENT_SECRET);;

  ipcMain.on('isAuthenticated', (event, arg) => {
    event.reply(`isAuthenticated-${arg.now}`, {isAuthenticated: isAuthenticated});
  });

  ipcMain.on('authenticate', (event, arg) => {
    authenticate().then((client) => {
      isAuthenticated = true;
      oauth2Client = client;
      event.reply(`authenticate-${arg.now}`, {isAuthenticated: true});
    }, (err) => {
      isAuthenticated = false;
      console.log(err);
      event.reply(`authenticate-${arg.now}`, {error: `Unable to authenticate with Google`, message: err, stack: err.stack})
    })
  });

  ipcMain.on('membershipLevels', (event, arg) => {
    getMembershipLevels(oauth2Client).then((data) => {
      console.log(data);
      event.reply(`membershipLevels-${arg.now}`, data.items)
    }, (err) => {
      console.log(err);
      event.reply(`membershipLevels-${arg.now}`, {error: `Unable to get membership levels`, message: err, stack: err.stack})
    })
  });

  ipcMain.on('members', (event, arg) => {
    getMembers(oauth2Client, arg.data).then((data) => {
      event.reply(`members-${arg.now}`, data.items);
    }, (err) => {
      event.reply(`members-${arg.now}`, {error: `Unable to get memberships`, message: err, stack: err.stack})
    })
  });

  ipcMain.on('save', (event, arg) => {
    dialog.showSaveDialog(win, {defaultPath: 'members.csv'}).then((res) => {
      if (res.canceled || !res.filePath) {
        return
      }

      fs.writeFile(res.filePath, arg.data, (err) => {
        if (err) {
          console.log(err);
          event.reply(`save-${arg.now}`, {error: `Unable to save ${res.filePath}`, message: err, stack: err.stack});
        }
        event.reply(`save-${arg.now}`, res);
      });
    });

    ipcMain.on('state:read', (event, arg) => {
      event.reply(`state:read-${arg.now}`, config.state);
    });

    ipcMain.on('state:write', (event, arg) => {
      config.state = arg.data;
      storeConfig(config);
    });
  });

  ipcMain.on('openExternal', (event, arg) => {
    shell.openExternal(arg.data);
  });

  ipcMain.on('signout', (event, arg) => {
    isAuthenticated = false;
    delete oauth2Client;
    removeToken();

    event.reply(`signout-${arg.now}`, {});
  });

  getToken().then((token) => {
    oauth2Client.credentials = token;
    isAuthenticated = true;
  }, (error) => {
    console.log(error)
  }).finally(() => {
    win.loadFile('index.html')
  });
}

app.whenReady().then(main)

app.on('window-all-closed', () => {
  if (process.platform !== 'darwin') {
    app.quit()
  }
})

app.on('activate', () => {
  if (BrowserWindow.getAllWindows().length === 0) {
    main()
  }
})

/**
 * Authenticate with Google.
 *    Creates the OAuth2 client.
 *    Starts a webserver to accept the OAuth2 callback response.
 *    Opens an external URL to authenticate with Google (uses the user's default browser, or most recenetly used browser).
 *    Returns a Promise.
 */
const authenticate = () => {
  return new Promise((resolve, reject) => {
    const server = http.createServer((request, response) => {
      const params = url.parse(request.url, true);

      if (params && params.query && params.query.code) {
        response.writeHead(200);
        // todo: look into using the API "setAsDefaultProtocolClient".
        response.end('success! you can now close this window and return to the Tukija application.');
        server.close();

        oauth2Client.getToken(params.query.code, (err, token) => {
          if (err) {
            console.log(err);
            reject(err);
            return;
          }
          oauth2Client.credentials = token;
          storeToken(token);
          resolve(oauth2Client);
        });
      }
    });

    server.listen(() => {
      const address = server.address();
      oauth2Client = new OAuth2(CLIENT_ID, CLIENT_SECRET, `http://[::1]:${address.port}`);
      const authUrl = oauth2Client.generateAuthUrl({
        access_type: 'offline',
        scope: SCOPES
      });

      // open the Google authUrl with the user's default browser, outside of Electron.
      shell.openExternal(authUrl);
    });
  });
}

const getChannel = (client) => {
  return youtube.channels.list({
    part: 'id',
    mine: true,
    auth: client
  });
}

const getMembers = (client, level) => {
  // todo: check for nextToken.
  return youtube.members.list({
    part: 'snippet',
    maxResults: 1000,
    hasAccessToLevel: level,
    auth: client
  });
}

const getMembershipLevels = (client) => {
  return youtube.membershipsLevels.list({
    auth: client,
    part: 'id, snippet'
  })
}

const storeToken = (token) => {
  return writeFile(TOKEN_PATH, JSON.stringify(token));
}

const removeToken = () => {
  return deleteFile(TOKEN_PATH);
}

const storeConfig = (config) => {
  return writeFile(CONFIG_PATH, JSON.stringify(config));
}

const getToken = () => {
  return new Promise((resolve, reject) => {
    readFile(TOKEN_PATH).then((data) => {
      resolve(JSON.parse(data));
    }, reject)
  });
}

const getConfig = () => {
  return new Promise((resolve, reject) => {
    readFile(CONFIG_PATH).then((data) => {
      resolve(JSON.parse(data));
    }, reject)
  });
}

const readFile = (file) => {
  return new Promise((resolve, reject) => {
    try {
      fs.readFile(file, (err, data) => {
        if (err) {
          console.log(err);
          reject(err);
          return;
        }
        resolve(data);
      });
    } catch (e) {
      reject(e)
    } finally {
      // nothing
    }
  });
}

const writeFile = (file, data) => {
  return new Promise((resolve, reject) => {
    fs.writeFile(file, data, (err) => {
      if (err) {
        console.log(err);
        reject(err);
        return;
      }
      resolve();
    });
  });
}

const deleteFile = (file) => {
  return new Promise((resolve, reject) => {
    fs.unlink(file, (err) => {
      if (err) {
        console.log(err);
        reject(err);
        return;
      }
      resolve();
    });
  });
}

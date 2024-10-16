const LOCAL_STORAGE_KEY = "last_used_auth_connection";

export class AuthManager {
  lastUsedConnection: string | null = null;
  //   selectedConnection: string | null = null;

  constructor() {
    this.lastUsedConnection = this.getLastUsedConnection();
    if (this.lastUsedConnection) {
      this.setLastUsedConnection(this.lastUsedConnection);
    } else {
      this.setLastUsedConnection(null);
    }
  }

  //   setSelectedConnection(connection: string | null) {
  //     this.selectedConnection = connection;
  //   }

  //   getSelectedConnection(): string | null {
  //     return this.selectedConnection;
  //   }

  setLastUsedConnection(connection: string | null) {
    if (connection) {
      localStorage.setItem(LOCAL_STORAGE_KEY, connection);
    } else {
      localStorage.removeItem(LOCAL_STORAGE_KEY);
    }
    this.lastUsedConnection = connection;
  }

  getLastUsedConnection(): string | null {
    return localStorage.getItem(LOCAL_STORAGE_KEY);
  }

  hasLastUsedConnection(): boolean {
    return this.getLastUsedConnection() !== null;
  }
}

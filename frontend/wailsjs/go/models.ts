export namespace main {
	
	export class DBNode {
	    name: string;
	    owner: string;
	    size: string;
	    current: boolean;
	
	    static createFrom(source: any = {}) {
	        return new DBNode(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.owner = source["owner"];
	        this.size = source["size"];
	        this.current = source["current"];
	    }
	}
	export class SchemaNode {
	    name: string;
	
	    static createFrom(source: any = {}) {
	        return new SchemaNode(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	    }
	}
	export class StatusSnapshot {
	    state: string;
	    port: number;
	    user: string;
	    password: string;
	    database: string;
	    host: string;
	    connectionUri: string;
	    jdbcUrl: string;
	    psqlCommand: string;
	    envBlock: string;
	    initialized: boolean;
	    dataDir: string;
	    binDir: string;
	    binariesPresent: boolean;
	
	    static createFrom(source: any = {}) {
	        return new StatusSnapshot(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.state = source["state"];
	        this.port = source["port"];
	        this.user = source["user"];
	        this.password = source["password"];
	        this.database = source["database"];
	        this.host = source["host"];
	        this.connectionUri = source["connectionUri"];
	        this.jdbcUrl = source["jdbcUrl"];
	        this.psqlCommand = source["psqlCommand"];
	        this.envBlock = source["envBlock"];
	        this.initialized = source["initialized"];
	        this.dataDir = source["dataDir"];
	        this.binDir = source["binDir"];
	        this.binariesPresent = source["binariesPresent"];
	    }
	}
	export class StorageInfo {
	    dataDir: string;
	    exists: boolean;
	    sizeBytes: number;
	    sizeHuman: string;
	    fileCount: number;
	    binDir: string;
	    logFile: string;
	
	    static createFrom(source: any = {}) {
	        return new StorageInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.dataDir = source["dataDir"];
	        this.exists = source["exists"];
	        this.sizeBytes = source["sizeBytes"];
	        this.sizeHuman = source["sizeHuman"];
	        this.fileCount = source["fileCount"];
	        this.binDir = source["binDir"];
	        this.logFile = source["logFile"];
	    }
	}
	export class TableNode {
	    name: string;
	    type: string;
	    rows: number;
	
	    static createFrom(source: any = {}) {
	        return new TableNode(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.type = source["type"];
	        this.rows = source["rows"];
	    }
	}

}


export namespace modbus {

	export class BitRow {
	    address: number;
	    displayAddress: number;
	    raw: string;
	    value: boolean;
	    timestamp: string;
	    quality: string;

	    static createFrom(source: any = {}) {
	        return new BitRow(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.address = source["address"];
	        this.displayAddress = source["displayAddress"];
	        this.raw = source["raw"];
	        this.value = source["value"];
	        this.timestamp = source["timestamp"];
	        this.quality = source["quality"];
	    }
	}
	export class ConnectRequest {
	    host: string;
	    port: number;
	    unitId: number;
	    timeoutMs: number;
	    retries: number;

	    static createFrom(source: any = {}) {
	        return new ConnectRequest(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.host = source["host"];
	        this.port = source["port"];
	        this.unitId = source["unitId"];
	        this.timeoutMs = source["timeoutMs"];
	        this.retries = source["retries"];
	    }
	}
	export class ConnectionStatus {
	    connected: boolean;
	    host: string;
	    port: number;
	    unitId: number;
	    timeoutMs: number;
	    retries: number;
	    lastError: string;
	    lastResponseMs: number;

	    static createFrom(source: any = {}) {
	        return new ConnectionStatus(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.connected = source["connected"];
	        this.host = source["host"];
	        this.port = source["port"];
	        this.unitId = source["unitId"];
	        this.timeoutMs = source["timeoutMs"];
	        this.retries = source["retries"];
	        this.lastError = source["lastError"];
	        this.lastResponseMs = source["lastResponseMs"];
	    }
	}
	export class FrameLog {
	    id: number;
	    timestamp: string;
	    direction: string;
	    transactionId: number;
	    protocolId: number;
	    length: number;
	    unitId: number;
	    functionCode: number;
	    payloadHex: string;
	    fullHex: string;
	    durationMs: number;
	    exceptionCode?: string;

	    static createFrom(source: any = {}) {
	        return new FrameLog(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.timestamp = source["timestamp"];
	        this.direction = source["direction"];
	        this.transactionId = source["transactionId"];
	        this.protocolId = source["protocolId"];
	        this.length = source["length"];
	        this.unitId = source["unitId"];
	        this.functionCode = source["functionCode"];
	        this.payloadHex = source["payloadHex"];
	        this.fullHex = source["fullHex"];
	        this.durationMs = source["durationMs"];
	        this.exceptionCode = source["exceptionCode"];
	    }
	}
	export class ReadBitsRequest {
	    functionCode: number;
	    address: number;
	    quantity: number;
	    addressBase: number;

	    static createFrom(source: any = {}) {
	        return new ReadBitsRequest(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.functionCode = source["functionCode"];
	        this.address = source["address"];
	        this.quantity = source["quantity"];
	        this.addressBase = source["addressBase"];
	    }
	}
	export class ReadBitsResponse {
	    functionCode: number;
	    address: number;
	    quantity: number;
	    rows: BitRow[];
	    durationMs: number;

	    static createFrom(source: any = {}) {
	        return new ReadBitsResponse(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.functionCode = source["functionCode"];
	        this.address = source["address"];
	        this.quantity = source["quantity"];
	        this.rows = this.convertValues(source["rows"], BitRow);
	        this.durationMs = source["durationMs"];
	    }

		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class ReadRegistersRequest {
	    functionCode: number;
	    address: number;
	    quantity: number;
	    dataType: string;
	    byteOrder: string;
	    addressBase: number;

	    static createFrom(source: any = {}) {
	        return new ReadRegistersRequest(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.functionCode = source["functionCode"];
	        this.address = source["address"];
	        this.quantity = source["quantity"];
	        this.dataType = source["dataType"];
	        this.byteOrder = source["byteOrder"];
	        this.addressBase = source["addressBase"];
	    }
	}
	export class RegisterRow {
	    address: number;
	    displayAddress: number;
	    registerCount: number;
	    rawHex: string;
	    uint16: number;
	    int16: number;
	    float32?: number;
	    binary: string;
	    value: any;
	    dataType: string;
	    timestamp: string;
	    quality: string;

	    static createFrom(source: any = {}) {
	        return new RegisterRow(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.address = source["address"];
	        this.displayAddress = source["displayAddress"];
	        this.registerCount = source["registerCount"];
	        this.rawHex = source["rawHex"];
	        this.uint16 = source["uint16"];
	        this.int16 = source["int16"];
	        this.float32 = source["float32"];
	        this.binary = source["binary"];
	        this.value = source["value"];
	        this.dataType = source["dataType"];
	        this.timestamp = source["timestamp"];
	        this.quality = source["quality"];
	    }
	}
	export class ReadRegistersResponse {
	    functionCode: number;
	    address: number;
	    quantity: number;
	    rows: RegisterRow[];
	    durationMs: number;

	    static createFrom(source: any = {}) {
	        return new ReadRegistersResponse(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.functionCode = source["functionCode"];
	        this.address = source["address"];
	        this.quantity = source["quantity"];
	        this.rows = this.convertValues(source["rows"], RegisterRow);
	        this.durationMs = source["durationMs"];
	    }

		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

	export class WriteMultipleCoilsRequest {
	    address: number;
	    values: boolean[];

	    static createFrom(source: any = {}) {
	        return new WriteMultipleCoilsRequest(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.address = source["address"];
	        this.values = source["values"];
	    }
	}
	export class WriteMultipleRegistersRequest {
	    address: number;
	    values: number[];

	    static createFrom(source: any = {}) {
	        return new WriteMultipleRegistersRequest(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.address = source["address"];
	        this.values = source["values"];
	    }
	}
	export class WritePreview {
	    functionCode: number;
	    address: number;
	    quantity: number;
	    value: number;
	    values: string[];
	    host: string;
	    port: number;
	    unitId: number;
	    registerHex: string;
	    pduHex: string;
	    frameHex: string;

	    static createFrom(source: any = {}) {
	        return new WritePreview(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.functionCode = source["functionCode"];
	        this.address = source["address"];
	        this.quantity = source["quantity"];
	        this.value = source["value"];
	        this.values = source["values"];
	        this.host = source["host"];
	        this.port = source["port"];
	        this.unitId = source["unitId"];
	        this.registerHex = source["registerHex"];
	        this.pduHex = source["pduHex"];
	        this.frameHex = source["frameHex"];
	    }
	}
	export class WriteSingleCoilRequest {
	    address: number;
	    value: boolean;

	    static createFrom(source: any = {}) {
	        return new WriteSingleCoilRequest(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.address = source["address"];
	        this.value = source["value"];
	    }
	}
	export class WriteSingleRegisterRequest {
	    address: number;
	    value: number;
	    dataType: string;
	    byteOrder: string;

	    static createFrom(source: any = {}) {
	        return new WriteSingleRegisterRequest(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.address = source["address"];
	        this.value = source["value"];
	        this.dataType = source["dataType"];
	        this.byteOrder = source["byteOrder"];
	    }
	}

}

export namespace profiles {

	export class Profile {
	    version: number;
	    name: string;
	    host: string;
	    port: number;
	    unitId: number;
	    timeoutMs: number;
	    retries: number;
	    addressBase: number;
	    dataType: string;
	    byteOrder: string;

	    static createFrom(source: any = {}) {
	        return new Profile(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.version = source["version"];
	        this.name = source["name"];
	        this.host = source["host"];
	        this.port = source["port"];
	        this.unitId = source["unitId"];
	        this.timeoutMs = source["timeoutMs"];
	        this.retries = source["retries"];
	        this.addressBase = source["addressBase"];
	        this.dataType = source["dataType"];
	        this.byteOrder = source["byteOrder"];
	    }
	}

}


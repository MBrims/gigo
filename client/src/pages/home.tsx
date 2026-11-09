import React, { useRef, useState } from 'react';
import { FormatBytes } from '../helpers';
import { Link } from 'react-router';

type InputFile = File | null | undefined;

interface FileInputProperties {
    StoredFile: InputFile | undefined;
    SetStoredFile: React.Dispatch<React.SetStateAction<InputFile>>;
}

function FileInput(properties: FileInputProperties) {
    const inputFile = useRef<HTMLInputElement | null>(null);

    const onDivClick = () => {
        inputFile?.current?.click();
    };

    const onFileChange = () => {
        properties.SetStoredFile(inputFile?.current?.files?.item(0));
    };

    type DivDragEvent = React.DragEvent<HTMLDivElement>;
    const onDragOver = (ev: DivDragEvent) => {
        ev.preventDefault();
    };

    const onDragLeave = (ev: DivDragEvent) => {
        ev.preventDefault();
    };

    const onDrop = (ev: DivDragEvent) => {
        ev.preventDefault();

        // Checking for nothing worthwhile
        if ((ev?.dataTransfer?.files?.length ?? 0) == 0) {
            return;
        }

        // Set the stored file here.
        properties.SetStoredFile(ev.dataTransfer.files.item(0));
    };

    return (
        <>
            <div className="drop-zone centered-div" onClick={onDivClick} onDragOver={onDragOver} onDragLeave={onDragLeave} onDrop={onDrop}>
                Drop File Here.
            </div>
            <input type='file' ref={inputFile} style={{ display: 'none' }} onChange={onFileChange} />
            <p style={{ alignSelf: 'center', textAlign: 'left' }}>
                File: {properties.StoredFile?.name ?? ''}<br></br>
                Size: {FormatBytes(properties.StoredFile?.size ?? 0)}
            </p>
        </>
    );
}

function StoreOnServer(props: { StoreOnServer: boolean, SetStoreOnServer: React.Dispatch<React.SetStateAction<boolean>> }) {
    const inputStoreOnServer = useRef<HTMLInputElement | null>(null);

    const onChange = () => {
        props.SetStoreOnServer(inputStoreOnServer?.current?.checked ?? true)
    }

    return (
        <div className="centered-div funny-lil-checkbox">
            <div className="check" id="store-server-div">
                <input id="store-server-checkbox" type='checkbox' ref={inputStoreOnServer} checked={props.StoreOnServer} onChange={onChange} />
                <label htmlFor="store-server-checkbox" />
            </div>
            <label htmlFor="store-server-div">Store on Server</label>
        </div>
    )
}

function TEMP_NAME(properties: { StoreOnServer: boolean, StoredFile: InputFile }) {
    // If we aren't uploading or don't have a file, just return.
    if (!properties.StoreOnServer || !properties.StoredFile) {
        return (
            <></>
        );
    }

    const uploadFile = () => {
        if (!properties.StoredFile) {
            return;
        }

        fetch("/upload", {
            method: "POST",
            headers: {
                "content-type": properties.StoredFile?.type,
                "content-disposition": "attachment; filename=" + properties.StoredFile?.name
            },
            body: properties.StoredFile,
        }).then(async (r) => {
            if (!r.ok) {
                alert(await r.text());
            }
            else {
                console.log(r);
                const str = new TextDecoder().decode(await r.bytes());
                console.log(str);

                alert(str);
                //const bytes = await r.bytes();
                //alert(Number(bytes));
            }
        })
    };

    return (
        <button className='upload-button' onClick={uploadFile}>Upload!</button>
    );
}

function ToDownloads() {
    return (
        <Link to="/download">To Downloads </Link>
    );
}

const Home = () => {
    const [storedFile, setStoredFiled] = useState<InputFile>(null);
    const [storeOnServer, setStoreOnServer] = useState(true);

    return (
        <div className="content">
            <h1>Add File</h1>
            <FileInput StoredFile={storedFile} SetStoredFile={setStoredFiled} />
            <StoreOnServer StoreOnServer={storeOnServer} SetStoreOnServer={setStoreOnServer} />
            <TEMP_NAME StoreOnServer={storeOnServer} StoredFile={storedFile} />
            <ToDownloads />
        </div>
    );
};

export default Home;

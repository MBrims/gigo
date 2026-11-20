/**
 * Page for handling the downloading of a file. This should be the landing page from a given URL, or allow the user to enter the credentials manually.
 */

import React, { useEffect, useState, type JSX } from "react";

// I got sick of writing them out.
function GrabQueryInfo() {
    const urlParams = new URLSearchParams(window.location.search);
    return urlParams.get('file');
}

// Specifically for downloading the given blob.
function DownloadFileHelper(name: string, blob: Blob) {
    const blobUrl = window.URL.createObjectURL(blob);
    const downloadElement = document.createElement('a');
    downloadElement.download = name;
    downloadElement.href = blobUrl;
    downloadElement.hidden = true;

    // The usual song and dance...
    document.body.appendChild(downloadElement);
    downloadElement.click();
    document.body.removeChild(downloadElement);
}

function DownloadFileServer(fileId: string | undefined) {
    if (!fileId) {
        return;
    }

    fetch("/file?file=" + fileId, {
        method: 'GET',
        headers: { 'Content-Type': 'application/octet-stream' }
    }).then(r => {

        // For the time being we'll just download the file directly.
        if (r.ok) {
            r.blob().then(b => {
                const disp = r.headers.get("Content-Disposition")
                const suggestName = disp?.split(';').find(s => s.includes("filename"))?.split('=')[1] ?? "";
                DownloadFileHelper(suggestName, b)
            })
        }
    });
}

function Query() {
    const queryInfo = GrabQueryInfo();
    return (
        <h1>{queryInfo ? "Found query: " + queryInfo : "No query found."}</h1>
    );
}

function HashDisplay() {
    const hash = window.location.hash;
    if (hash.length == 0) {
        return <h1>No hash found.</h1>;
    }

    // Remove the leading #
    return <p>Found hash: {hash.slice(1)}</p>
}

const QueryInputField = (ops: { fileQuery: string, setFileQuery: React.Dispatch<React.SetStateAction<string>> }) => {
    const [inputValue, setInputValue] = useState<string>(ops.fileQuery);

    useEffect(() => {
        setInputValue(ops.fileQuery);
    }, [ops.fileQuery]);

    return (
        <div>
            <input type="text" placeholder="Enter Hash" value={inputValue} onChange={e => setInputValue(e.target.value)} />
            <button onClick={() => ops.setFileQuery(inputValue)} >Query</button>
        </div>
    )
}

function GetHash() {
    // For some reason it includes the #, so slice that off.
    return window.location.hash.slice(1);
}

const Download = () => {
    // The file we want to query.
    const [fileQuery, setFileQuery] = useState<string>(GetHash()); //GetHash());

    // The JSON representation of the file object.
    const [fileInfo, setFileInfo] = useState<object>();

    // When the hash in the window changes, we want to update.
    useEffect(() => {
        const onHashChange = () => {
            setFileQuery(GetHash());
        }
        window.addEventListener('hashchange', onHashChange);
        return () => window.removeEventListener('hashchange', onHashChange);
    }, [])

    // The file stream from the server.
    // const [fileStream, setFileStream] = useState<ReadableStream<Uint8Array<ArrayBuffer>>>();

    // Initial check to see if we have included a query in the hash.
    // useEffect(() => {
    //     const possibleHash = GetHash();
    //     if (possibleHash.length > 0) {
    //         setFileQuery(possibleHash);
    //     }
    // }, []);

    // Update the info if the query changes.
    useEffect(() => {
        if (fileQuery?.length ?? 0 > 0) { // This would send an empty string on first load otherwise.
            window.location.hash = "#" + fileQuery; // For the sake of consistency, we'll set the hash.
            fetch("/fileInfo?file=" + fileQuery,
                {
                    method: 'GET',
                    headers: {
                        "Content-Type": "application/json",
                    }
                }
            ).then(async (r) => {
                if (r.ok) {
                    setFileInfo(await r.json());
                }
                else {
                    console.log(r.status, ":", r.statusText);
                    setFileInfo(undefined) // Getting rid of it. Ideally this should instead have an error or something.
                }
            });
        }
        else {
            window.location.hash = ''; // Just remove it.
        }
    }, [fileQuery]);

    const fileInfoValues: JSX.Element[] = []

    if (fileInfo) {
        //@ts-expect-error We're grabbing the keys already.
        fileInfoValues = Object.keys(fileInfo).map(key => (<li>{key} : {fileInfo[key].toString()}</li>))
    }

    return (
        <>
            <QueryInputField fileQuery={fileQuery} setFileQuery={setFileQuery} />
            <div hidden={fileInfoValues.length == 0}>
                <ul>
                    {fileInfoValues}
                </ul>
                <button onClick={() => DownloadFileServer(fileQuery)} >Download</button>
            </div>
        </>
    )
}

const Download_old = () => {
    const [loadedQuery, setLoadedQuery] = useState<boolean>(false);
    const [loadedData, setLoadedData] = useState<string>();

    if (!loadedQuery) {
        const queryInfo = GrabQueryInfo();
        fetch("/download?file=" + queryInfo).then((r) => {
            if (r.ok) {
                // Split it and then join it back together.
                const suggestedFileName = r.headers.get('Content-Disposition')?.split("=").slice(1).join("=");
                console.log(suggestedFileName);

                r.text().then((v) => {
                    setLoadedData(v);
                    setLoadedQuery(true);
                });
            }
        });
    }

    if (loadedQuery) {
        return (
            <div>
                <p>{loadedData}</p>
                <HashDisplay />
                <a href="/">Back to Home</a>
            </div>
        )
    }

    return (
        <div>
            <Query />
            <HashDisplay />
            <a href="/">Back to Home</a>
        </div>
    );
};

export default Download;

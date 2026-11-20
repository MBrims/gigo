/**
 * This page should just have some stats about the running server; total space taken up sorta deal, # of files, etc.
 */

import { useEffect, useState, type JSX } from "react"

function statusOk(status: number) {
    return (status >= 200 && status < 300)
}

type statsResponseT = { status: number, resp: any }

// Returns an array with the first element being the initial status code.
async function requestStats(): Promise<statsResponseT> {
    const resp = await fetch("/statistics")
    if (resp.ok) {
        const json = await resp.json();
        console.log(json);

        return { status: resp.status, resp: json }
    }

    return { status: resp.status, resp: resp.statusText }
}



const Stats = () => {
    const [loadedStats, setLoadedStats] = useState<object>();

    // We only want to run this once.
    useEffect(() => {
        requestStats().then(r => {
            if (statusOk(r.status)) {
                setLoadedStats(r.resp);
            }
        });
    }, []);

    const statValues: JSX.Element[] = []

    if (loadedStats) {
        //@ts-expect-error We're grabbing the keys already.
        statValues = Object.keys(loadedStats).map(key => (<li>{key} : {loadedStats[key].toString()}</li>))

        // Object.keys(loadedStats).map(key => `${key} : ${loadedStats[key]}`)

        // for (const kv in Object.entries(loadedStats)) {
        //     statValues.push(<li>{kv[0]} : {kv[1]}</li>)
        // }
    }

    return (
        <>
        <h1>Stats</h1>
        <p>Some statistics about the server. Currently it's just slapping the JSON down as is.</p>
        <hr/>
            <ul>
                {statValues}
            </ul>
        </>
    )
}

export default Stats;
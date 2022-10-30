async function getLastWeekDownloadStats() {
    const response = await fetch('/api/v1/stats/downloads/past-week');
    const data = await response.json();
    return data;
}

async function getLastWeekIndexStats() {
    const response = await fetch('/api/v1/stats/indexes/past-week');
    const data = await response.json();
    return data;
}

(async function(){
    let downloadStats = await getLastWeekDownloadStats();
    if(!downloadStats) {
        downloadStats = [];
    }


    const downloadStatsCanvas = document.getElementById('downloadStats');

    const downloadsChart = new Chart(downloadStatsCanvas, {
        type: 'line',
        data: {
            labels: downloadStats.map(stat => stat.date) || [],
            datasets: [{
                label: 'Download Stats',
                data: downloadStats.map(stat => stat.count) || [],
                borderWidth: 1
            }]
        }
    }); 

    let indexStats = await getLastWeekIndexStats();
    if(!indexStats) {
        indexStats = [];
    }

    const indexStatsCanvas = document.getElementById('indexStats');
    const indexesChart = new Chart(indexStatsCanvas, {
        type: 'line',
        data: {
            labels: indexStats.map(stat => stat.date) || [],
            datasets: [{
                label: 'Index Stats',
                data: indexStats.map(stat => stat.count) || [],
                borderWidth: 1
            }]
        }
    }); 
})()
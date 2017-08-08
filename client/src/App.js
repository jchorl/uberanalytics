import React, { Component } from 'react';

function nanosToString(nanos) {
    var hours = Math.floor(nanos / 3.6e+12);
    var minutes = Math.floor((nanos % 3.6e+12) / 6e+10);
    var seconds = ((nanos % 6e+10) / 1e+9).toFixed(0);
    return hours + ":" + minutes + ":" + (seconds < 10 ? '0' : '') + seconds;
}

class App extends Component {
    constructor(props) {
        super(props);
        this.state = {
            authd: false,
            totalOnTrip: '',
            totalWaiting: '',
            averageTimePerTrip: '',
        };
    }

    componentWillMount() {
        const headers = new Headers();
        headers.set('Accept', 'application/json');

        fetch('/api/auth', {
            method: 'GET',
            credentials: 'include',
            headers
        })
        .then(resp => {
            if (resp.ok) {
                this.setState({ authd: true });
                this.fetchStats();
            }
        })
        .catch(e => console.error(e));
    }

    fetchStats = () => {
        const headers = new Headers();
        headers.set('Accept', 'application/json');

        fetch('/api/stats', {
            method: 'GET',
            credentials: 'include',
            headers
        })
        .then(resp => resp.json())
        .then(parsed => {
            this.setState({
                totalOnTrip: nanosToString(parsed.totalOnTrip),
                totalWaiting: nanosToString(parsed.totalWaiting),
                averageTimePerTrip: nanosToString(parsed.totalOnTrip / parsed.count),
            });
        })
        .catch(e => console.error(e));
    }

    goToAuth = () => {
        window.location = "https://login.uber.com/oauth/v2/authorize?client_id=6p5LewpobI8T3_yuNIX7KiAU--XWUH0t&response_type=code&scope=history";
    }

    render() {
        const {
            authd,
            totalOnTrip,
            totalWaiting,
            averageTimePerTrip,
        } = this.state;

        return (
                <div>
                    {
                    authd
                    ? (
                    <div>
                        <div>
                            Total on trips: { totalOnTrip }
                        </div>
                        <div>
                            Total waiting: { totalWaiting }
                        </div>
                        <div>
                            Average trip time: { averageTimePerTrip }
                        </div>
                    </div>
                    )
                    : <button onClick={ this.goToAuth }>Auth</button>
                    }
                </div>
                );
    }
}

export default App;

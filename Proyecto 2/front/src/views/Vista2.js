import React from 'react'
import { Button } from 'semantic-ui-react'
import Navbar from '../components/Navbar'

function Vista2() {


    const vista = () => {
        window.location.href = '/vista3'
    }

    return (
        <div>
            <Navbar />

            Vista2
            <br />
            <br />

            <Button inverted color='green' onClick={vista}>
                vista3
            </Button>

        </div>
    )
}

export default Vista2
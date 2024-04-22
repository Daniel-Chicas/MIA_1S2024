import React from 'react'
import { Button } from 'semantic-ui-react'
import Navbar from '../components/Navbar'

function Vista3() {
    const vista = () => {
        window.location.href = '/'
    }
    return (
        <div>

            <Navbar />
            Vista3
            <br />
            <br />
            <Button inverted color='green' onClick={vista}>
                vista1
            </Button>

        </div>
    )
}

export default Vista3
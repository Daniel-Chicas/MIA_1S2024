import React, { useState } from 'react'
import { Form, TextArea, Button } from 'semantic-ui-react'
import Navbar from '../components/Navbar'

function Vista1() {
    const [text, setText] = useState('')

    const cambioTextArea = (event) => {
        setText(event.target.value)
    }

    const enviar = () => {
        console.log(text)
    }

    const vista = () => {
        window.location.href = '/vista2'
    }

    return (
        <>
            <Navbar />

            <Form>
                <TextArea rows={2} placeholder='Tell us more' value={text} onChange={cambioTextArea} />
            </Form>

            <Button inverted color='red' onClick={enviar}>
                imprimir
            </Button>

            <Button inverted color='green' onClick={vista}>
                vista2
            </Button>
        </>
    )
}

export default Vista1
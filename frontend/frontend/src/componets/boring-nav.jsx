import { useState } from 'react'
import './boring.css'

function Boring() {
  return (
    <>
        <div className="nav">  //buton navigation instead of termial. how do u comment in react lol
            <button>About me</button>
            <button>Projects</button>
            <button>Contact</button>
            <button>BTS</button>
        </div>
    </>
  )
}

export default Boring
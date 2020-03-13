import React from 'react';
import SwaggerUI from "swagger-ui-react"
import "swagger-ui-react/swagger-ui.css"
import { RedocStandalone } from 'redoc';

function App({url, mode}) {
  if (mode === "redoc") {
    return (<RedocStandalone specUrl={url}/>)
  }
  return (<SwaggerUI url={url} />)
}

export default App;

import React from 'react';

export default function OS() {
  var default_os = "mac";  // most common OS per install history
  // let mobile folks get the default binary
  if (navigator.platform.indexOf("Mac") > -1) default_os="mac";
  else if (navigator.platform.indexOf("Win") > -1) default_os="win";
  else if (navigator.platform.indexOf("Linux") > -1) default_os="linux";
  return default_os;
}

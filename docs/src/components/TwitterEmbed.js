import React, { useEffect } from 'react';

const TwitterEmbed = ({ tweetId }) => {
  useEffect(() => {
    const script = document.createElement('script');
    script.src = "https://platform.twitter.com/widgets.js";
    script.async = true;
    document.body.appendChild(script);
  }, []);

  return <blockquote className="twitter-tweet">
            <a href={`https://twitter.com/user/status/${tweetId}`}></a>
         </blockquote>;
};

export default TwitterEmbed;
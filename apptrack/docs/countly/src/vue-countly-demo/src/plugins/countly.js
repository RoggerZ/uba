import Countly from 'countly-sdk-web';

export default {
  install(app, options) {
    // Initialize Countly
    Countly.init({
      app_key: options.appKey || '2963dcc87fe3f000725a16c9ada61f707ca546dd',
      url: options.url || 'http://local.count.ly:8380',
      debug: options.debug !== undefined ? options.debug : true
    });

    // Enable features (matching official guide)
    Countly.track_sessions();
    Countly.track_pageview();
    Countly.track_clicks();
    Countly.track_scrolls();      // Added: track scroll depth
    Countly.track_errors();
    Countly.track_links();        // Added: track link clicks
    Countly.track_forms();        // Added: track form interactions
    Countly.collect_from_forms(); // Added: collect form data

    // Add to global properties
    app.config.globalProperties.$countly = {
      // Record event
      recordEvent(eventName, segmentation = {}, count = 1) {
        Countly.add_event({
          key: eventName,
          count: count,
          segmentation: segmentation
        });
      },

      // Start timed event
      startEvent(eventName) {
        Countly.start_event(eventName);
      },

      // End timed event
      endEvent(eventName, segmentation = {}) {
        Countly.end_event({
          key: eventName,
          segmentation: segmentation
        });
      },

      // Set user details
      setUserDetails(userDetails) {
        Countly.user_details(userDetails);
      },

      // Set custom user property
      setCustomProperty(key, value) {
        Countly.user_details({
          custom: {
            [key]: value
          }
        });
      },

      // Track page view
      trackPageView(pageName) {
        Countly.track_pageview(pageName);
      },

      // Log error
      logError(error) {
        Countly.log_error(error);
      },

      // Add crash log
      addCrashLog(log) {
        Countly.add_log(log);
      }
    };

    console.log('✓ Countly plugin installed successfully');
  }
};

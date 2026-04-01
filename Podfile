# Uncomment the next line to define a global platform for your project
platform :macos, '10.15'

target 'xNetFuture' do
  # Comment the next line if you don't want to use dynamic frameworks
  use_frameworks!
  
  pod 'MMKV', '~> 1.2.16'

  # Pods for xNetFuture

end

post_install do |installer|
  installer.pods_project.targets.each do |target|
    target.build_configurations.each do |config|
      config.build_settings['MACOSX_DEPLOYMENT_TARGET'] = '10.15'
    end
  end
end

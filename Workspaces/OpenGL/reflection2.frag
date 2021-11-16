#version 330
in vec2 UV;
layout(location = 0) out vec4 frag_color;

//Uniforms
uniform vec3 clearColor;
uniform vec3 camPos;
uniform vec3 lookAt;


uniform float shadowK;
uniform float shadowEpsilon;
//Rendering things
uniform int MAX_MARCHING_STEPS;
uniform float MAX_DIST = 100.0;
//Control over scene
uniform int DBDRAW;

uniform vec3 lightPos;
uniform vec3 spherePos;
uniform float sphereRad;

uniform vec3 boxPos;
uniform float boxRotationZ;
uniform float boxRotationX;

uniform float reflectMix;

uniform sampler2D image;

//Constants
const float MIN_DIST = 0.0;
const float EPSILON = 0.0001;
const float CAMRADIUS= .15;

uniform float ambient;






struct Material {
   vec3 color;
   float reflectivity;
};

Material mats[4];

void makeMats(){
mats[1].color = vec3(1,0,1);
mats[1].reflectivity = 0;

mats[0].color = vec3(1,1,1);
mats[0].reflectivity = reflectMix;

mats[2].color = vec3(.4,.4,.4);
mats[2].reflectivity = 0;
    
//"Camera"
mats[3].color = vec3(.1);
mats[3].reflectivity = 0;
}






vec3 rotateZ(vec3 p, float a){
    return vec3(
        cos(a)*p.x-sin(a)*p.y,
        cos(a)*p.y+sin(a)*p.x,
        p.z);
}

vec3 rotateX(vec3 p, float a){
    return vec3(
        p.x,
        cos(a)*p.y-sin(a)*p.z,
        cos(a)*p.z+sin(a)*p.y
        );
}


struct sdResult{
    float distance;
    int materialID;
};

sdResult sdRound(sdResult sd, float radius){
    sd.distance-=radius;
    return sd;
}

sdResult sdSmoothUnion(sdResult a, sdResult b, float k){
    float d1 = a.distance;
    float d2 = b.distance;
    float h = clamp( 0.5 + 0.5*(d2-d1)/k, 0.0, 1.0 );
    sdResult res;
    res.distance = mix( d2, d1, h ) - k*h*(1.0-h);
    res.materialID = h>.5 ? a.materialID : b.materialID;
    return res;
}
sdResult sdMin(sdResult a, sdResult b){
    if (a.distance<b.distance){
        return a;
    }
    return b;
}
sdResult sdMax(sdResult a, sdResult b){
    if (a.distance>b.distance){
        return a;
    }
    return b;
}

sdResult sdSub(sdResult a, sdResult b){
    a.distance = -a.distance;
    if (a.distance>b.distance){
        return a;
    }
    return b;
}

sdResult sdSphere(vec3 samplePoint, float radius, int matID) {
    sdResult res;
    res.distance = length(samplePoint) - radius;
    res.materialID = matID;
    return res;
}

sdResult sdBox( vec3 p, vec3 b, int matID)
{
  vec3 q = abs(p) - b;
  sdResult res;
  res.distance = length(max(q,0.0)) + min(max(q.x,max(q.y,q.z)),0.0);
  res.materialID = matID;
  return res;
}

sdResult sdXYPlane(vec3 p, float z, int matID){
    sdResult res;
    res.distance = abs(p.z-z);
    res.materialID = matID;
    return res;
}

sdResult sceneSDF(vec3 samplePoint) {
    sdResult res;
    res = sdMin( 
        sdMin( 
            sdXYPlane(samplePoint,-2,2),
            sdSphere(samplePoint-camPos, CAMRADIUS,3)
        ),
        sdMin(
          sdSphere(samplePoint-spherePos,sphereRad,1),
          sdBox(rotateX(rotateZ(samplePoint-boxPos,boxRotationZ),boxRotationX), vec3(1,1,1), 0)
        )
        );
    return res;
}

struct DistRes {
   float dist;
   int steps;
};

DistRes shortestDistanceToSurface(vec3 ro, vec3 rd, float start, float end) {
    DistRes result;
    //result.closestDist = 
    float depth = start;
    int i;
    for (i = 0; i < MAX_MARCHING_STEPS; i++) {
        sdResult res = sceneSDF(ro + depth * rd); 
        float dist = res.distance;
        if (dist < EPSILON) {
			result.dist = depth;
            result.steps = i;
			return result;
        }
        depth += dist;
        if (depth >= end) {
            result.dist = end;
            result.steps = i;
            return result;
        }
    }
    result.dist = end;        
    result.steps = i;
    return result;
    
}
struct ShadowRes {
   float closest;
   int steps;
};

ShadowRes softshadow(vec3 ro, vec3 rd, float mint, float maxt, float k)
{
    ShadowRes result;
    result.steps = 0; 
    result.closest = 1.0;
    for( float t=mint; t<maxt; )
    {        
        result.steps++;
        sdResult distInfo = sceneSDF(ro + rd*t);
        float h = distInfo.distance;
        if( h<0.001 ){
            result.closest = 0;
            return result;
        }
        result.closest = min( result.closest, k*h/t );
        t += h;

    }
    return result;
}


vec3 estimateNormal(vec3 p) {
    return normalize(vec3(
        sceneSDF(vec3(p.x + EPSILON, p.y, p.z)).distance - sceneSDF(vec3(p.x - EPSILON, p.y, p.z)).distance,
        sceneSDF(vec3(p.x, p.y + EPSILON, p.z)).distance - sceneSDF(vec3(p.x, p.y - EPSILON, p.z)).distance,
        sceneSDF(vec3(p.x, p.y, p.z  + EPSILON)).distance - sceneSDF(vec3(p.x, p.y, p.z - EPSILON)).distance
    ));
}


vec3 phong(vec3 p, vec3 eye ){
    //Blinn Phong shading        
    vec3 N = estimateNormal(p);
    vec3 L = normalize(lightPos - p);
    vec3 V = normalize(eye - p);
    vec3 R = normalize(reflect(-L, N));
    
    float dotLN = dot(L, N);
    float dotRV = dot(R, V);
    
    vec3 k_s = vec3(1); //Specular color
    float alpha = 10; //Shinyness
    vec3 lightCol;
    vec3 lightIntensity = vec3(.4);
    if (dotLN < 0.0) {
        // Light not visible from this point on the surface
        lightCol = vec3(0.0, 0.0, 0.0);
    } else if (dotRV < 0.0) {
        // Light reflection in opposite direction as viewer, apply only diffuse
        // component
        lightCol = lightIntensity  * dotLN;
    } else {
        lightCol = lightIntensity * dotLN + k_s * pow(dotRV, alpha);
    }

    return lightCol;
}



//uv should be x[-1,1] and y[-1,1]

vec3 rayDirection(float fieldOfView, vec2 uv) {
    vec2 xy = uv;
    float z = 1 / tan(radians(fieldOfView) / 2.0);
    return normalize(vec3(xy, -z));
}


mat4 viewMatrix(vec3 eye, vec3 center, vec3 up) {
	vec3 f = normalize(center - eye);
	vec3 s = normalize(cross(f, up));
	vec3 u = cross(s, f);
	return mat4(
		vec4(s, 0.0),
		vec4(u, 0.0),
		vec4(-f, 0.0),
		vec4(0.0, 0.0, 0.0, 1)
	);
}





vec3 normalToEnv(vec3 n){
    vec2 sample = vec2(atan(n.y, n.x)/6.282, 1-(n.z+1)/2);
    return texture(image,sample).xyz;
}

void updateMats(vec3 p){
    bool map = floor(mod(p.x,2)) == 0 ^^ floor(mod(p.y,2)) == 0;
    mats[2].color = mix(vec3(.8), vec3(.7), float(map));
}

float makeShadow(vec3 p, vec3 n){
    vec3 sstart = p + n*shadowEpsilon;
    ShadowRes shadow = softshadow(sstart, normalize(lightPos-sstart), 0 , length(sstart-lightPos), shadowK);
    return shadow.closest;
}

void main() {

    float aspect = 8.0/6.0;
    vec2 uv = UV/vec2(aspect, 1);

    
	vec3 dir = rayDirection(45.0, uv);
    vec3 eye = camPos;
    
        
    mat4 viewToWorld = viewMatrix(eye, lookAt, vec3(0.0, 1.0, 0.0));
    vec3 worldDir = (viewToWorld * vec4(dir, 0.0)).xyz;
    
    
    DistRes Res = shortestDistanceToSurface(eye, worldDir, MIN_DIST+CAMRADIUS+0.05, MAX_DIST);
    
    float dist = Res.dist;
    
    if (dist > MAX_DIST - EPSILON) {
        // Didn't hit anything
        vec3 cColor = normalToEnv(worldDir);   
        frag_color = vec4(cColor, 1.0);    
		return;
    }


    makeMats();

    // The closest point on the surface to the eyepoint along the view ray
    vec3 p = eye + dist * worldDir;

    //Genereate texture
    updateMats(p);
            
    int matID = sceneSDF(p).materialID;
   
   
    vec3 normal = estimateNormal(p);
   //Shadow start to avoid immediatly dieing
    float closest = makeShadow(p, normal);

    Material mat = mats[matID];
    vec3 color = mat.color;
    //vec3 lightCol = phong(p, eye);
    //color += lightCol;
    color = color * clamp(closest+ambient, 0, 1);



    //Second Bounce
    vec3 newDir = reflect(worldDir, normal);
    DistRes newRes = shortestDistanceToSurface(p, newDir, MIN_DIST+shadowEpsilon, MAX_DIST);    
    float newDist = newRes.dist;
    
    
    
    vec3 newP = p + newDist * newDir + normal*shadowEpsilon;

    vec3 newCol;
    if (newDist<MAX_DIST-EPSILON){
        vec3 normal2 = estimateNormal(p);
    
        float closest2 = makeShadow(newP, normal2);

        updateMats(newP);
    
    
        Material newMat = mats[sceneSDF(newP).materialID];
        
        newCol = newMat.color;
        
        newCol = newCol * clamp(closest2+ambient, 0, 1);
    } else {
        newCol=normalToEnv(newDir);
    }
    

    color = mix(color, newCol, mat.reflectivity);

    color = pow(color,vec3(1.9));

    if (DBDRAW == 0){
        color = vec3(float(Res.steps)/float(MAX_MARCHING_STEPS),0,0);
    } else if (DBDRAW ==1){
        color = abs(normal);
    } else if (DBDRAW==2){
        color = abs(newDir);
    } else if (DBDRAW==3){
        color=vec3(matID/4.0);
    } else if (DBDRAW==4){
        color = (fract(p));
    } else if (DBDRAW == 5){
        color = abs(worldDir);
    }

    frag_color = vec4(color, 1.0);


}
